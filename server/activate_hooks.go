package main

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API)

	botID, ensureBotError := p.Helpers.EnsureBot(&model.Bot{
		Username:    "channelexport",
		DisplayName: "Channel Export Bot",
		Description: "A bot account created by the channel export plugin.",
	})
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure bot.")
	}

	p.botID = botID

	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated. This is the plugin's last chance to use
// the API, and the plugin will be terminated shortly after this invocation.
func (p *Plugin) OnDeactivate() error {
	return nil
}
