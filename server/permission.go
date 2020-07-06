package main

import (
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v5/model"
)

// showEmailAddress determines if the given user has access to email addresses within the system.
func showEmailAddress(client *pluginapi.Wrapper, userID string) bool {
	var showEmailAddress bool
	if client.User.HasPermissionTo(userID, model.PERMISSION_MANAGE_SYSTEM) {
		showEmailAddress = true
	} else {
		config := client.Configuration.GetConfig()
		if config != nil && config.PrivacySettings.ShowEmailAddress != nil {
			showEmailAddress = *config.PrivacySettings.ShowEmailAddress
		}
	}

	return showEmailAddress
}
