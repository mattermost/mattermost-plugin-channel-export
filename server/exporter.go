package main

import (
	"io"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

type Exporter interface {
	Export(channel *model.Channel, writer io.Writer) error
}

// ExportedPost contains all the information from a post needed in
// an export, with all the relevant information already resolved
type ExportedPost struct {
	CreateAt     int64
	UserID       string
	UserEmail    string
	UserType     string
	UserName     string
	ID           string
	ParentPostID string
	Message      string
	Type         string
}

func ToExportedPost(client *pluginapi.Client, post *model.Post) (*ExportedPost, error) {
	user, err := client.User.Get(post.UserId)
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
		UserID:       post.UserId,
		UserEmail:    user.Email,
		UserType:     userType,
		UserName:     user.Nickname,
		ID:           post.Id,
		ParentPostID: post.ParentId,
		Message:      post.Message,
		Type:         postType,
	}, nil
}
