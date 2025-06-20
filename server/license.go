// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost/server/public/model"
	originalapi "github.com/mattermost/mattermost/server/public/pluginapi"
)

func isLicensed(_ *model.License, api *pluginapi.Wrapper) bool {
	return originalapi.IsE20LicensedOrDevelopment(api.Configuration.GetConfig(), api.System.GetLicense())
}
