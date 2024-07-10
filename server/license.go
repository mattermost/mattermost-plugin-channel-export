package main

import (
	"github.com/mattermost/mattermost-server/v6/model"

	originalapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
)

func isLicensed(_ *model.License, api *pluginapi.Wrapper) bool {
	return originalapi.IsE20LicensedOrDevelopment(api.Configuration.GetConfig(), api.System.GetLicense())
}
