// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package pluginapi

import (
	"io"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"

	pluginapi "github.com/mattermost/mattermost/server/public/pluginapi"
)

// Channel is an interface declaring only the functions from
// mattermost-plugin-api ChannelService that are used in this plugin
type Channel interface {
	Get(channelID string) (*model.Channel, error)
	GetDirect(userID1, userID2 string) (*model.Channel, error)
}

// File is an interface declaring only the functions from
// mattermost-plugin-api FileService that are used in this plugin
type File interface {
	Upload(content io.Reader, fileName, channelID string) (*model.FileInfo, error)
}

// Log is an interface declaring only the functions from
// mattermost-plugin-api LogService that are used in this plugin
type Log interface {
	Error(message string, keyValuePairs ...interface{})
}

// Post is an interface declaring only the functions from
// mattermost-plugin-api PostService that are used in this plugin
type Post interface {
	CreatePost(post *model.Post) error
	GetPostsForChannel(channelID string, page, PerPage int) (*model.PostList, error)
}

// SlashCommand is an interface declaring only the functions from
// mattermost-plugin-api SlashCommandService that are used in this plugin
type SlashCommand interface {
	Register(command *model.Command) error
}

// User is an interface declaring only the functions from
// mattermost-plugin-api UserService that are used in this plugin
type User interface {
	Get(userID string) (*model.User, error)
	HasPermissionTo(userID string, permission *model.Permission) bool
	HasPermissionToChannel(userID, channelID string, permission *model.Permission) bool
}

// System is an interface declaring only the functions from
// mattermost-plugin-api SystemService that are used in this plugin
type System interface {
	GetLicense() *model.License
}

// Configuration is an interface declaring only the functions from
// mattermost-plugin-api ConfigurationService that are used in this plugin
type Configuration interface {
	GetConfig() *model.Config
}

type Cluster interface {
	NewMutex(key string) (ClusterMutex, error)
}

// Wrapper is a wrapper over the mattermost-plugin-api layer, defining
// interfaces implemented by that package, that are also mockable
type Wrapper struct {
	Channel       Channel
	File          File
	Log           Log
	Post          Post
	SlashCommand  SlashCommand
	User          User
	System        System
	Configuration Configuration
	Cluster       Cluster
}

// CustomWrapper builds a Wrapper with the implementations of the different
// interfaces passed
func CustomWrapper(
	channel Channel,
	file File,
	log Log,
	post Post,
	slashCommand SlashCommand,
	user User,
	system System,
	configuration Configuration,
	cluster Cluster,
) *Wrapper {
	return &Wrapper{
		Channel:       channel,
		File:          file,
		Log:           log,
		Post:          post,
		SlashCommand:  slashCommand,
		User:          user,
		System:        system,
		Configuration: configuration,
		Cluster:       cluster,
	}
}

// Wrap wraps a plugin.API with the mattermost-plugin-api layer, interfaced by
// this package
func Wrap(client *pluginapi.Client, api plugin.API) *Wrapper {
	return CustomWrapper(
		&client.Channel,
		&client.File,
		&client.Log,
		&client.Post,
		&client.SlashCommand,
		&client.User,
		&client.System,
		&client.Configuration,
		&ClusterService{api: api},
	)
}
