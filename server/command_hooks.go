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
	go p.exportChannel(args)

	text := "Exporting the channel. Wait while we create the file for you. Our bot will send you a message when we are finished."
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         text,
	}
}

func (p *Plugin) exportChannel(args *model.CommandArgs) {
	type channelChan struct {
		channel *model.Channel
		err     error
	}

	// Create a Direct Message channel between the user and the plugin's bot
	dmChan := make(chan channelChan, 1)
	go func() {
		channel, err := p.client.Channel.GetDirect(args.UserId, p.botID)
		dmChan <- channelChan{channel, err}
		close(dmChan)
	}()

	// Retrieve the channel to export
	toExportChan := make(chan channelChan, 1)
	go func() {
		channel, err := p.client.Channel.Get(args.ChannelId)
		toExportChan <- channelChan{channel, err}
		close(toExportChan)
	}()

	dmChanResponse := <-dmChan
	channelDM, err := dmChanResponse.channel, dmChanResponse.err
	if err != nil {
		p.client.Log.Error("Unable to create a Direct Message channel between the bot and the user.",
			"Bot ID", p.botID, "User ID", args.UserId, "Error", err)

		text := "There was an error trying to create a Direct Message channel between you and our bot."
		p.client.Post.SendEphemeralPost(args.UserId, &model.Post{
			UserId:    p.botID,
			ChannelId: args.ChannelId,
			Message:   text,
		})
		return
	}

	toExportChanResponse := <-toExportChan
	channelToExport, err := toExportChanResponse.channel, toExportChanResponse.err
	if err != nil {
		p.client.Log.Error("Unable to retrieve the channel to export",
			"Channel ID", args.ChannelId, "Error", err)

		text := "There was an error trying to retrieve the channel to export: \n" + err.Error()
		p.client.Post.SendEphemeralPost(args.UserId, &model.Post{
			UserId:    p.botID,
			ChannelId: args.ChannelId,
			Message:   text,
		})
		return
	}

	// Send an empty JSON file for now. The actual implementation will come later.
	fileName := fmt.Sprintf("%d_%s.json", time.Now().Unix(), channelToExport.Name)
	fileContents := strings.NewReader("{}")
	file, err := p.client.File.Upload(fileContents, fileName, channelDM.Id)
	if err != nil {
		p.client.Log.Error("Unable to upload the exported file to the channel.",
			"Channel ID", channelDM.Id, "Error", err)

		text := "There was an error exporting the file."
		p.client.Post.SendEphemeralPost(args.UserId, &model.Post{
			UserId:    p.botID,
			ChannelId: channelDM.Id,
			Message:   text,
		})
		return
	}

	text := fmt.Sprintf("Channel ~%s exported:", channelToExport.Name)
	p.client.Post.CreatePost(&model.Post{
		UserId:    p.botID,
		ChannelId: channelDM.Id,
		Message:   text,
		FileIds:   []string{file.Id},
	})
}
