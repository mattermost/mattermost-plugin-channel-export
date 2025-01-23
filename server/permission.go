// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v6/model"
)

// showEmailAddress determines if the given user has access to email addresses within the system.
func showEmailAddress(client *pluginapi.Wrapper, userID string) bool {
	if client.User.HasPermissionTo(userID, model.PermissionManageSystem) {
		return true
	}

	config := client.Configuration.GetConfig()
	if config != nil && config.PrivacySettings.ShowEmailAddress != nil {
		return *config.PrivacySettings.ShowEmailAddress
	}

	return false
}
