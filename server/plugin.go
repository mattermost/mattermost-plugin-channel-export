package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	pluginAPIWrapper "github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	client *pluginAPIWrapper.Wrapper
	botID  string

	// router is the plugin's root HTTP handler
	router *mux.Router

	// makeChannelPostsIterator is a factory function for iterating over posts
	makeChannelPostsIterator func(*model.Channel, bool) PostIterator
}

const (
	botUsername    = "channelexport"
	botDisplayName = "Channel Export Bot"
	botDescription = "A bot account created by the channel export plugin."
)

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate() error {
	client := pluginapi.NewClient(p.API)
	p.client = pluginAPIWrapper.Wrap(client)
	pluginapi.ConfigureLogrus(logrus.New(), client)

	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    botUsername,
		DisplayName: botDisplayName,
		Description: botDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot.")
	}

	p.botID = botID

	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	p.router = mux.NewRouter()
	p.makeChannelPostsIterator = func(channel *model.Channel, showEmailAddress bool) PostIterator {
		return channelPostsIterator(p.client, channel, showEmailAddress)
	}
	registerAPI(p.router, p.client, p.makeChannelPostsIterator)

	return nil
}

// ServeHTTP handles requests to /plugins/com.mattermost.plugin-incident-response
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}
