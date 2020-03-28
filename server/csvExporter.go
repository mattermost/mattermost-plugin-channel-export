package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type CSVExporter struct {
	client *pluginapi.Client
}

func (e *CSVExporter) postToCSVLine(post *ExportedPost) []string {
	line := []string{
		strconv.FormatInt(post.CreateAt, 10),
		post.UserID,
		post.UserEmail,
		post.UserType,
		post.UserName,
		post.ID,
		post.ParentPostID,
		post.Message,
		post.Type,
	}
	return line
}

func (e *CSVExporter) Export(channel *model.Channel, writer io.Writer) error {
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
		postList, err := e.client.Post.GetPostsForChannel(channel.Id, page, perPage)

		if err != nil {
			return fmt.Errorf("Unable to fetch page %d (with %d posts) from channel %s: %w", 0, 2, channel.Id, err)
		}

		for _, key := range postList.Order {
			post := postList.Posts[key]
			// Ignore posts that have been edited; exporting only what's visible in the channel
			if post.OriginalId != "" {
				continue
			}

			exportedPost, err := ToExportedPost(e.client, post)
			if err != nil {
				return fmt.Errorf("Unable to export post: %w", err)
			}

			line := e.postToCSVLine(exportedPost)
			csvWriter.Write(line)
		}

		if len(postList.Posts) < perPage {
			break
		}
	}

	csvWriter.Flush()

	return nil
}
