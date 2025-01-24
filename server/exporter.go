// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/utils"
)

const (
	userTypeRegular = "user"
	userTypeBot     = "bot"
	userTypeSystem  = "system"

	PerPage = 500
)

// PostIterator returns the next batch of posts when called
type PostIterator func() ([]*ExportedPost, error)

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
// called, a new batch of posts from the channel, chronologically ordered
// (most recent first), until all posts have been consumed.
func channelPostsIterator(client *pluginapi.Wrapper, channel *model.Channel, showEmailAddress bool) PostIterator {
	usersCache := make(map[string]*model.User)
	page := 0
	return func() ([]*ExportedPost, error) {
		postList, err := client.Post.GetPostsForChannel(channel.Id, page, PerPage)
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

			exportedPost, err := toExportedPost(client, post, showEmailAddress, usersCache)
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
func toExportedPost(client *pluginapi.Wrapper, post *model.Post, showEmailAddress bool, usersCache map[string]*model.User) (*ExportedPost, error) {
	user, ok := usersCache[post.UserId]
	if !ok {
		newUser, err := client.User.Get(post.UserId)
		if err != nil {
			return nil, errors.Wrap(err, "failed retrieving post's author information")
		}

		usersCache[post.UserId] = newUser
		user = newUser
	}

	userType := userTypeRegular
	if user.IsBot {
		userType = userTypeBot
	}

	postType := "message"
	if post.Type != "" {
		postType = post.Type
		userType = userTypeSystem
	}

	exportedPost := &ExportedPost{
		CreateAt:  utils.TimeFromMillis(post.CreateAt).UTC(),
		UserID:    post.UserId,
		UserEmail: "",
		UserType:  userType,
		UserName:  user.Username,
		ID:        post.Id,
		Message:   post.Message,
		Type:      postType,
	}

	if showEmailAddress {
		exportedPost.UserEmail = user.Email
	}

	return exportedPost, nil
}
