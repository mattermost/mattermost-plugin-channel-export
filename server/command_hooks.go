package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"

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

	go func() {
		fileName := fmt.Sprintf("%d_%s.json", time.Now().Unix(), channelToExport.Name)
		fileContents, err := p.exportChannel(channelToExport)
		if err != nil {
			p.client.Post.CreatePost(&model.Post{
				UserId:    p.botID,
				ChannelId: channelDM.Id,
				Message:   fmt.Sprintf("An error occurred exporting channel ~%s.", channelToExport.Name),
			})

			return
		}

		file, err := p.uploadExportedChannelTo(fileName, fileContents, args.UserId)
		if err != nil {
			p.client.Post.CreatePost(&model.Post{
				UserId:    p.botID,
				ChannelId: channelDM.Id,
				Message:   fmt.Sprintf("An error occurred uploading the exported channel ~%s.", channelToExport.Name),
			})

			return
		}

		p.client.Post.CreatePost(&model.Post{
			UserId:    p.botID,
			ChannelId: channelDM.Id,
			Message:   fmt.Sprintf("Channel ~%s exported:", channelToExport.Name),
			FileIds:   []string{file.Id},
		})
	}()

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: fmt.Sprintf("Exporting ~%s. @%s will send you a direct message when the export is ready.",
			channelToExport.Name, botUsername),
	}
}

// Export an empty JSON file for now. The actual implementation will come later.
func (p *Plugin) exportChannel(channel *model.Channel) (io.Reader, error) {
	fileContents := strings.NewReader("{}")
	return fileContents, nil
}

func (p *Plugin) uploadExportedChannelTo(fileName string, fileContents io.Reader, receiverID string) (*model.FileInfo, error) {
	file, err := p.client.File.Upload(fileContents, fileName, receiverID)
	if err != nil {
		p.client.Log.Error("unable to upload the exported file to the channel",
			"Channel ID", receiverID, "Error", err)
		return nil, fmt.Errorf("unable to upload the exported file")
	}

	return file, nil
}
