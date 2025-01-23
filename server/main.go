// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/mattermost/mattermost-server/v6/plugin"
)

func main() {
	plugin.ClientMain(&Plugin{})
}
