// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
)

func TestShowEmailAddress(t *testing.T) {
	trueValue := true
	falseValue := false

	t.Run("system administrator", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)

		mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
		mockFile := mock_pluginapi.NewMockFile(mockCtrl)
		mockLog := mock_pluginapi.NewMockLog(mockCtrl)
		mockPost := mock_pluginapi.NewMockPost(mockCtrl)
		mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
		mockUser := mock_pluginapi.NewMockUser(mockCtrl)
		mockSystem := mock_pluginapi.NewMockSystem(mockCtrl)
		mockConfiguration := mock_pluginapi.NewMockConfiguration(mockCtrl)
		mockCluster := mock_pluginapi.NewMockCluster(mockCtrl)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		userID := "user_id"

		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(true).Times(1)
		assert.True(t, showEmailAddress(mockAPI, userID))
	})

	t.Run("not system administrator, show email address", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)

		mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
		mockFile := mock_pluginapi.NewMockFile(mockCtrl)
		mockLog := mock_pluginapi.NewMockLog(mockCtrl)
		mockPost := mock_pluginapi.NewMockPost(mockCtrl)
		mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
		mockUser := mock_pluginapi.NewMockUser(mockCtrl)
		mockSystem := mock_pluginapi.NewMockSystem(mockCtrl)
		mockConfiguration := mock_pluginapi.NewMockConfiguration(mockCtrl)
		mockCluster := mock_pluginapi.NewMockCluster(mockCtrl)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		userID := "user_id"

		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})
		assert.True(t, showEmailAddress(mockAPI, userID))
	})

	t.Run("not system administrator, hide email address", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)

		mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
		mockFile := mock_pluginapi.NewMockFile(mockCtrl)
		mockLog := mock_pluginapi.NewMockLog(mockCtrl)
		mockPost := mock_pluginapi.NewMockPost(mockCtrl)
		mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
		mockUser := mock_pluginapi.NewMockUser(mockCtrl)
		mockSystem := mock_pluginapi.NewMockSystem(mockCtrl)
		mockConfiguration := mock_pluginapi.NewMockConfiguration(mockCtrl)
		mockCluster := mock_pluginapi.NewMockCluster(mockCtrl)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		userID := "user_id"

		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &falseValue,
			},
		})
		assert.False(t, showEmailAddress(mockAPI, userID))
	})
}
