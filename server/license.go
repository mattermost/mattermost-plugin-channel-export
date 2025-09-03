// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost/server/public/model"
	originalapi "github.com/mattermost/mattermost/server/public/pluginapi"
)

func isLicensed(_ *model.License, api *pluginapi.Wrapper) bool {
	license := api.System.GetLicense()
	config := api.Configuration.GetConfig()

	if originalapi.IsConfiguredForDevelopment(config) {
		return true
	}

	return originalapi.IsE20LicensedOrDevelopment(config, license) && license != nil && license.SkuShortName != "entry"
}
