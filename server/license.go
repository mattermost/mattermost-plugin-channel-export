package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

func isLicensed(license *model.License) bool {
	if license == nil {
		return false
	}

	if license.Features == nil {
		return false
	}

	if license.Features.FutureFeatures == nil {
		return false
	}

	if !*license.Features.FutureFeatures {
		return false
	}

	return true
}
