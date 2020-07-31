package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
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
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case exportCommandTrigger:
		return p.executeCommandExport(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("Unknown command: " + args.Command),
		}, nil
	}
}

func (p *Plugin) executeCommandExport(args *model.CommandArgs) *model.CommandResponse {
	license := p.client.System.GetLicense()
	if !isLicensed(license) {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "The channel export plugin requires a valid E20 license.",
		}
	}

	channelToExport, err := p.client.Channel.Get(args.ChannelId)
	if err != nil {
		p.client.Log.Error("unable to retrieve the channel to export",
			"Channel ID", args.ChannelId, "Error", err)

		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Unable to retrieve the channel to export.",
		}
	}

	channelDM, err := p.client.Channel.GetDirect(args.UserId, p.botID)
	if err != nil {
		p.client.Log.Error("unable to create a direct message channel between the bot and the user",
			"Bot ID", p.botID, "User ID", args.UserId, "Error", err)

		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         fmt.Sprintf("An error occurred trying to create a direct message channel between you and @%s.", botUsername),
		}
	}

	exporter := CSV{}
	fileName := exporter.FileName(channelToExport.Name)

	exportError := errors.New("failed to export channel")

	logger := logrus.WithFields(logrus.Fields{
		"user_id":    args.UserId,
		"channel_id": channelToExport.Id,
	})

	exportedFileReader, exportedFileWriter := io.Pipe()
	go func() {
		err := exporter.Export(p.makeChannelPostsIterator(channelToExport, showEmailAddress(p.client, args.UserId)), exportedFileWriter)
		if err != nil {
			logger.WithError(err).Warn("failed to export channel")

			_ = exportedFileWriter.CloseWithError(errors.Wrap(exportError, err.Error()))

			err = p.client.Post.CreatePost(&model.Post{
				UserId:    p.botID,
				ChannelId: channelDM.Id,
				Message:   fmt.Sprintf("An error occurred exporting channel ~%s.", channelToExport.Name),
			})
			if err != nil {
				logger.WithError(err).Warn("failed to post message about failure to export channel")
			}

			return
		}

		exportedFileWriter.Close()
	}()

	go func() {
		file, err := p.uploadFileTo(fileName, exportedFileReader, channelDM.Id)
		if err != nil {
			logger.WithError(err).Warn("failed to upload exported channel")

			// Post the upload error only if the exporter did not do it before
			if !errors.Is(err, exportError) {
				err = p.client.Post.CreatePost(&model.Post{
					UserId:    p.botID,
					ChannelId: channelDM.Id,
					Message:   fmt.Sprintf("An error occurred uploading the exported channel ~%s.", channelToExport.Name),
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
			Message:   fmt.Sprintf("Channel ~%s exported:", channelToExport.Name),
			FileIds:   []string{file.Id},
		})
		if err != nil {
			logger.WithError(err).Warn("failed to post message about exported channel")
		}
	}()

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: fmt.Sprintf("Exporting ~%s. @%s will send you a direct message when the export is ready.",
			channelToExport.Name, botUsername),
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
