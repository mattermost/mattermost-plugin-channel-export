package apiwrapper

import (
	"io"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
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
}

// Wrapper is a wrapper over the mattermost-plugin-api layer, defining
// interfaces implemented by that package, that are also mockable
type Wrapper struct {
	Channel      Channel
	File         File
	Log          Log
	Post         Post
	SlashCommand SlashCommand
	User         User
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
) *Wrapper {
	return &Wrapper{
		Channel:      channel,
		File:         file,
		Log:          log,
		Post:         post,
		SlashCommand: slashCommand,
		User:         user,
	}
}

// Wrap wraps a plugin.API with the mattermost-plugin-api layer, interfaced by
// this package
func Wrap(api plugin.API) *Wrapper {
	underlyingWrapper := pluginapi.NewClient(api)

	return CustomWrapper(
		&underlyingWrapper.Channel,
		&underlyingWrapper.File,
		&underlyingWrapper.Log,
		&underlyingWrapper.Post,
		&underlyingWrapper.SlashCommand,
		&underlyingWrapper.User,
	)
}
