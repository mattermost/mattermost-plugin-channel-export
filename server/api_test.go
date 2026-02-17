// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/wiggin77/merror"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
)

func setupAPI(t *testing.T, mockAPI *pluginapi.Wrapper, now time.Time, userID, _ /*channelID*/ string, pluginConfiguration *configuration) string {
	router := mux.NewRouter()
	p := Plugin{
		router:        router,
		client:        mockAPI,
		configuration: pluginConfiguration,
	}

	err := registerAPI(&p, makeTestPostsIterator(t, now))
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate what the Mattermost server would normally do after validating a token.
		if userID != "" {
			r.Header.Add("Mattermost-User-ID", userID)
		}

		router.ServeHTTP(w, r)
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	return ts.URL
}

func TestHandler(t *testing.T) {
	trueValue := true
	falseValue := false

	t.Run("unauthorized", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		address := setupAPI(t, mockAPI, time.Now(), "", "channel_id", nil)
		client := NewClient(address)

		err := client.ExportChannel(io.Discard, "channel_id", FormatCSV)
		require.EqualError(t, err, "failed with status code 401")
	})

	t.Run("missing e20 license and no Testing nor Developer modes enabled", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		address := setupAPI(t, mockAPI, time.Now(), "user_id", "channel_id", nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(nil).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				EnableTesting:   &falseValue,
				EnableDeveloper: &falseValue,
			},
		}).Times(1)

		err := client.ExportChannel(io.Discard, "channel_id", FormatCSV)
		require.EqualError(t, err, "the channel export plugin requires a valid Enterprise license.")
	})

	t.Run("missing e20 license with Testing and Developer modes enabled", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(nil).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				EnableTesting:   &trueValue,
				EnableDeveloper: &trueValue,
			},
		}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.NoError(t, err)

		expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

		require.Equal(t, expected, buffer.String())
	})

	t.Run("missing channel_id", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		address := setupAPI(t, mockAPI, time.Now(), "user_id", "channel_id", nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)

		err := client.ExportChannel(io.Discard, "", FormatCSV)
		require.EqualError(t, err, "missing channel_id parameter")
	})

	t.Run("missing format", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		address := setupAPI(t, mockAPI, time.Now(), "user_id", "channel_id", nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)

		err := client.ExportChannel(io.Discard, "channel_id", "")
		require.EqualError(t, err, "missing format parameter")
	})

	t.Run("unsupported format", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		address := setupAPI(t, mockAPI, time.Now(), "user_id", "channel_id", nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)

		err := client.ExportChannel(io.Discard, "channel_id", "pdf2")
		require.EqualError(t, err, "unsupported format parameter 'pdf2'")
	})

	t.Run("channel not found", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), "user_id", channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(nil, &model.AppError{StatusCode: http.StatusNotFound}).Times(1)

		err := client.ExportChannel(io.Discard, channelID, FormatCSV)
		require.EqualError(t, err, "channel 'channel_id' not found or user does not have permission")
	})

	t.Run("failed querying channel", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), "user_id", channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(nil, &model.AppError{StatusCode: http.StatusInternalServerError}).Times(1)

		err := client.ExportChannel(io.Discard, channelID, FormatCSV)
		require.EqualError(t, err, "channel 'channel_id' not found or user does not have permission")
	})

	t.Run("no permission to channel", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(false).Times(1)

		err := client.ExportChannel(io.Discard, channelID, FormatCSV)
		require.EqualError(t, err, "channel 'channel_id' not found or user does not have permission")
	})

	t.Run("export with channel read permission, without access to email", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &falseValue,
			},
		}).Times(1)

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.NoError(t, err)

		expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

		require.Equal(t, expected, buffer.String())
	})

	t.Run("export with channel read permission, with access to email", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.NoError(t, err)

		expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

		require.Equal(t, expected, buffer.String())
	})

	t.Run("don't allow concurrent exports", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Do(func() {
			t.Log("about to sleep for 3s")
			time.Sleep(time.Second * 3)
		}).Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})

		countExec := 3
		merr := merror.New()
		wg := sync.WaitGroup{}

		for range countExec {
			wg.Go(func() {
				var buffer bytes.Buffer
				if err := client.ExportChannel(&buffer, channelID, FormatCSV); err != nil {
					merr.Append(err)
				} else {
					expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`
					require.Equal(t, expected, buffer.String())
				}
			})
		}

		wg.Wait()

		require.Equal(t, countExec-1, merr.Len())
		for _, err := range merr.Errors() {
			require.Equal(t, "a channel export is already running.", err.Error())
		}
	})

	t.Run("export when channel is archived and not visible", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID, DeleteAt: 1}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &falseValue,
			},
		}).Times(2)

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.EqualValues(t, "", buffer.String())

		expectedErr := fmt.Sprintf("channel '%s' is archived and not visible anymore", channelID)

		require.EqualError(t, err, expectedErr)
	})

	t.Run("export when channel is archived and visible", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, nil)
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID, DeleteAt: 1}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(true).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.Nil(t, err)

		expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`
		require.EqualValues(t, expected, buffer.String())
	})

	t.Run("no permissions", func(t *testing.T) {
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
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID, &configuration{
			EnableAdminRestrictions: true,
		})
		client := NewClient(address)
		client.SetToken("token")

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID, DeleteAt: 0}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionReadChannel).Return(true).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PermissionManageChannelRoles).Return(false).Times(1)
		mockUser.EXPECT().HasPermissionTo(userID, model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			TeamSettings: model.TeamSettings{
				ExperimentalViewArchivedChannels: &trueValue,
			},
		}).Times(2)

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.EqualValues(t, "", buffer.String())

		expectedErr := "user does not have permission to export channels"

		require.EqualError(t, err, expectedErr)
	})
}
