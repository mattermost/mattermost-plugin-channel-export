package main

import (
	"io"
	"time"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/utils"
	"github.com/pkg/errors"
)

// PostIterator returns the next batch of posts when called
type PostIterator func() ([]*ExportedPost, error)

// Exporter processes a list of posts and writes them to a writer
type Exporter interface {
	// FileName returns the name of the exported file, given the core name passed
	FileName(name string) string

	// Export processes the posts returned by nextPosts and exports them to writer
	Export(nextPosts PostIterator, writer io.Writer) error
}

// ExportedPost contains all the information from a post needed in
// an export, with all the relevant information already resolved
type ExportedPost struct {
	CreateAt     time.Time
	UserID       string
	UserEmail    string
	UserType     string
	UserName     string
	ID           string
	ParentPostID string
	Message      string
	Type         string
}

// channelPostsIterator returns a function that returns, every time it is
// called, a new batch of posts from the channel, chronollogically ordered
// (most recent first), until all posts have been consumed.
func channelPostsIterator(client *pluginapi.Wrapper, channel *model.Channel) PostIterator {
	usersCache := make(map[string]*model.User)
	page := 0
	perPage := 1000
	return func() ([]*ExportedPost, error) {
		postList, err := client.Post.GetPostsForChannel(channel.Id, page, perPage)
		if err != nil {
			return nil, err
		}

		exportedPostList := make([]*ExportedPost, 0, len(postList.Order))
		for _, key := range postList.Order {
			post := postList.Posts[key]

			// Ignore posts that have been edited; exporting only what's visible in the channel
			if post.OriginalId != "" {
				continue
			}

			exportedPost, err := toExportedPost(client, post, usersCache)
			if err != nil {
				return nil, errors.Wrap(err, "unable to export post")
			}

			exportedPostList = append(exportedPostList, exportedPost)
		}

		page++
		return exportedPostList, nil
	}
}

// toExportedPost resolves all the data from post that is needed in
// ExportedPost, as the user information and the type of message
func toExportedPost(client *pluginapi.Wrapper, post *model.Post, usersCache map[string]*model.User) (*ExportedPost, error) {
	user, ok := usersCache[post.UserId]
	if !ok {
		newUser, err := client.User.Get(post.UserId)
		if err != nil {
			return nil, errors.Wrap(err, "failed retrieving post's author information")
		}

		usersCache[post.UserId] = newUser
		user = newUser
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
		CreateAt:     utils.TimeFromMillis(post.CreateAt),
		UserID:       post.UserId,
		UserEmail:    user.Email,
		UserType:     userType,
		UserName:     user.Username,
		ID:           post.Id,
		ParentPostID: post.ParentId,
		Message:      post.Message,
		Type:         postType,
	}, nil
}
