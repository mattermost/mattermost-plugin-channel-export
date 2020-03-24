package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
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

	go func() {
		defer pipeWriter.Close()
		err := p.exportChannel(channelToExport, pipeWriter)
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

func (p *Plugin) exportChannel(channel *model.Channel, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	err := csvWriter.Write([]string{
		"Post Creation Time",
		"User Id",
		"User Email",
		"User Type",
		"User Name",
		"Post Id",
		"Parent Post Id",
		"Post Message",
		"Post Type",
	})

	if err != nil {
		return fmt.Errorf("Unable to create a CSV file: %w", err)
	}

	for page, perPage := 0, 100; ; page++ {
		// FIXME: Swap page and perPage parameters when https://github.com/mattermost/mattermost-server/pull/14125 is released
		postList, err := p.client.Post.GetPostsForChannel(channel.Id, page, perPage)

		if err != nil {
			return fmt.Errorf("Unable to fetch page %d (with %d posts) from channel %s: %w", 0, 2, channel.Id, err)
		}

		for _, key := range postList.Order {
			post := postList.Posts[key]
			// Ignore posts that have been edited; exporting only what's visible in the channel
			if post.OriginalId != "" {
				continue
			}

			exportedPost, err := p.exportedPost(post)
			if err != nil {
				return fmt.Errorf("Unable to export post: %w", err)
			}

			line := postToCSVLine(exportedPost)
			csvWriter.Write(line)
		}

		if len(postList.Posts) < perPage {
			break
		}
	}

	csvWriter.Flush()

	return nil
}

type ExportedPost struct {
	CreateAt     int64
	UserId       string
	UserEmail    string
	UserType     string
	UserName     string
	Id           string
	ParentPostId string
	Message      string
	Type         string
}

func (p *Plugin) exportedPost(post *model.Post) (*ExportedPost, error) {
	user, err := p.client.User.Get(post.UserId)
	if err != nil {
		return nil, errors.Wrap(err, "failed retrieving post's author information")
	}

	userType := "user"
	if user.IsBot {
		userType = "bot"
	}

	postType := "message"
	if post.Type != "" {
		postType = post.Type
		userType = "system"
	}

	return &ExportedPost{
		CreateAt:     post.CreateAt,
		UserId:       post.UserId,
		UserEmail:    user.Email,
		UserType:     userType,
		UserName:     user.Nickname,
		Id:           post.Id,
		ParentPostId: post.ParentId,
		Message:      post.Message,
		Type:         postType,
	}, nil
}

func postToCSVLine(post *ExportedPost) []string {
	line := []string{
		strconv.FormatInt(post.CreateAt, 10),
		post.UserId,
		post.UserEmail,
		post.UserType,
		post.UserName,
		post.Id,
		post.ParentPostId,
		post.Message,
		post.Type,
	}
	return line
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
