package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

const exportCommandTrigger = "export"

func (p *Plugin) registerCommands() error {
	if err := p.client.SlashCommand.Register(&model.Command{
		Trigger:          exportCommandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Export the current channel.",
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", exportCommandTrigger)
	}

	return nil
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand
// API.
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case exportCommandTrigger:
		return p.executeCommandExport(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: " + args.Command),
		}, nil
	}
}

func (p *Plugin) executeCommandExport(args *model.CommandArgs) *model.CommandResponse {
	// only allow one export at a time
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	if err := p.clusterMutex.LockWithContext(ctx); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "An export is already running.",
		}
	}
	var active bool
	defer func() {
		if !active {
			p.clusterMutex.Unlock()
		}
	}()

	license := p.client.System.GetLicense()
	if !isLicensed(license, p.client) {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "The channel export plugin requires a valid E20 license.",
		}
	}

	if !p.hasPermissionToExportChannel(args.UserId, args.ChannelId) {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "You do not have enough permissions to export this channel",
		}
	}

	channelToExport, err := p.client.Channel.Get(args.ChannelId)
	if err != nil {
		p.client.Log.Error("unable to retrieve the channel to export",
			"Channel ID", args.ChannelId, "Error", err)

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Unable to retrieve the channel to export.",
		}
	}

	var channelToExportName, exportSuccessMsg, exportProcessMsg string
	if channelToExport.Type == model.ChannelTypeOpen || channelToExport.Type == model.ChannelTypePrivate {
		channelToExportName = channelToExport.Name
		exportSuccessMsg = fmt.Sprintf("Channel ~%s exported:", channelToExportName)
		exportProcessMsg = "Exporting ~%s. @%s will send you a direct message when the export is ready."
	} else if channelToExport.Type == model.ChannelTypeGroup {
		channelToExportName = channelToExport.Name
		exportProcessMsg = "Exporting this GM channel ~%s. @%s will send you a direct message when the export is ready."
		team, err := p.API.GetTeam(args.TeamId)
		if err != nil {
			p.client.Log.Error("error occurred while getting the details of team for the channel.",
				"GM channel ID", args.ChannelId, "User ID", args.UserId, "Error", err)

			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("An error occurred trying to exporting the chat for %s GM channel", channelToExportName),
			}
		}
		link := fmt.Sprintf("%s/%s/messages/%s", args.SiteURL, team.Name, channelToExportName)
		exportSuccessMsg = fmt.Sprintf("GM Channel ~[%s](%s) exported:", channelToExportName, link)

	} else if channelToExport.Type == model.ChannelTypeDirect {
		DMUserName := strings.Split(channelToExport.Name, "__")[0]
		user, _ := p.client.User.Get(DMUserName)
		channelToExportName = user.Username
		exportSuccessMsg = fmt.Sprintf("DM with @%s exported:", channelToExportName)
		exportProcessMsg = "Exporting DM with @%s. @%s will send you a direct message when the export is ready."
	}

	channelDM, err := p.client.Channel.GetDirect(args.UserId, p.botID)
	if err != nil {
		p.client.Log.Error("unable to create a direct message channel between the bot and the user",
			"Bot ID", p.botID, "User ID", args.UserId, "Error", err)

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("An error occurred trying to create a direct message channel between you and @%s.", botUsername),
		}
	}

	exporter := CSV{}
	fileName := exporter.FileName(channelToExportName)

	exportError := errors.New("failed to export channel")

	logger := logrus.WithFields(logrus.Fields{
		"user_id":    args.UserId,
		"channel_id": channelToExport.Id,
	})

	exportedFileReader, exportedFileWriter := io.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(2)
	active = true

	go func() {
		defer wg.Done()
		err := exporter.Export(p.makeChannelPostsIterator(channelToExport, showEmailAddress(p.client, args.UserId)), exportedFileWriter)
		if err != nil {
			logger.WithError(err).Warn("failed to export channel")

			_ = exportedFileWriter.CloseWithError(errors.Wrap(exportError, err.Error()))

			err = p.client.Post.CreatePost(&model.Post{
				UserId:    p.botID,
				ChannelId: channelDM.Id,
				Message:   fmt.Sprintf("An error occurred exporting channel ~%s.", channelToExportName),
			})
			if err != nil {
				logger.WithError(err).Warn("failed to post message about failure to export channel")
			}

			return
		}

		exportedFileWriter.Close()
	}()

	go func() {
		defer wg.Done()
		file, err := p.uploadFileTo(fileName, exportedFileReader, channelDM.Id)
		if err != nil {
			logger.WithError(err).Warn("failed to upload exported channel")

			// Post the upload error only if the exporter did not do it before
			if !errors.Is(err, exportError) {
				err = p.client.Post.CreatePost(&model.Post{
					UserId:    p.botID,
					ChannelId: channelDM.Id,
					Message:   fmt.Sprintf("An error occurred uploading the exported channel ~%s.", channelToExportName),
				})
				if err != nil {
					logger.WithError(err).Warn("failed to post message about failure to upload exported channel")
				}
			}

			return
		}

		err = p.client.Post.CreatePost(&model.Post{
			UserId:    p.botID,
			ChannelId: channelDM.Id,
			Message:   exportSuccessMsg,
			FileIds:   []string{file.Id},
		})
		if err != nil {
			logger.WithError(err).Warn("failed to post message about exported channel")
		}
	}()

	// wait until both goroutines above are completed then mark exporter inactive
	go func() {
		defer p.clusterMutex.Unlock()
		wg.Wait()
	}()

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf(exportProcessMsg, channelToExportName, botUsername),
	}
}

func (p *Plugin) uploadFileTo(fileName string, contents io.Reader, channelID string) (*model.FileInfo, error) {
	file, err := p.client.File.Upload(contents, fileName, channelID)
	if err != nil {
		p.client.Log.Error("unable to upload the exported file to the channel",
			"Channel ID", channelID, "Error", err)
		return nil, errors.Wrap(err, "unable to upload the exported file")
	}

	return file, nil
}

func (p *Plugin) hasPermissionToExportChannel(userID, channelID string) bool {
	conf := p.getConfiguration()
	if conf.EnableAdminRestrictions {
		if !(p.client.User.HasPermissionToChannel(userID, channelID, model.PermissionManageChannelRoles) || p.client.User.HasPermissionTo(userID, model.PermissionManageSystem)) {
			return false
		}
	}
	return true
}
