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

	pipeReader, pipeWriter := io.Pipe()
	fileName := fmt.Sprintf("%d_%s.csv", time.Now().Unix(), channelToExport.Name)

	// TODO: Add logic to choose from different exporters when they are implemented
	exporter := CSVExporter{}

	go func() {
		defer pipeWriter.Close()

		err := exporter.Export(p.postIterator(channelToExport), pipeWriter)
		if err != nil {
			p.client.Post.CreatePost(&model.Post{
				UserId:    p.botID,
				ChannelId: channelDM.Id,
				Message:   fmt.Sprintf("An error occurred exporting channel ~%s.", channelToExport.Name),
			})

			return
		}
	}()

	go func() {
		file, err := p.uploadExportedChannelTo(fileName, pipeReader, args.UserId)
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

func (p *Plugin) uploadExportedChannelTo(fileName string, contents io.Reader, receiverID string) (*model.FileInfo, error) {
	file, err := p.client.File.Upload(contents, fileName, receiverID)
	if err != nil {
		p.client.Log.Error("unable to upload the exported file to the channel",
			"Channel ID", receiverID, "Error", err)
		return nil, fmt.Errorf("unable to upload the exported file")
	}

	return file, nil
}

func (p *Plugin) postIterator(channel *model.Channel) PostIterator {
	page := 0
	return func(perPage int) ([]*ExportedPost, error) {
		// FIXME: Swap page and perPage parameters when https://github.com/mattermost/mattermost-server/
		postList, err := p.client.Post.GetPostsForChannel(channel.Id, perPage, page)
		if err != nil {
			return nil, err
		}

		var exportedPostList []*ExportedPost
		for _, key := range postList.Order {
			post := postList.Posts[key]
			// Ignore posts that have been edited; exporting only what's visible in the channel
			if post.OriginalId != "" {
				continue
			}

			exportedPost, err := ToExportedPost(p.client, post)
			if err != nil {
				return nil, fmt.Errorf("Unable to export post: %w", err)
			}

			exportedPostList = append(exportedPostList, exportedPost)
		}

		page += 1
		return exportedPostList, nil
	}
}
