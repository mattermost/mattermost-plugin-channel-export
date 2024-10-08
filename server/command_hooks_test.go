package main

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

func setupPlugin(t *testing.T, mockAPI *pluginapi.Wrapper, now time.Time) (*Plugin, *plugin.Context) {
	clusterMutex, err := mockAPI.Cluster.NewMutex(KeyClusterMutex)
	require.NoError(t, err)

	return &Plugin{
		client:                   mockAPI,
		botID:                    "bot_id",
		makeChannelPostsIterator: makeTestPostsIterator(t, now),
		clusterMutex:             clusterMutex,
	}, &plugin.Context{}
}

func TestExecuteCommand(t *testing.T) {
	trueValue := true
	falseValue := false

	t.Run("unexpected trigger", func(t *testing.T) {
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

		plugin, pluginContext := setupPlugin(t, mockAPI, time.Now())

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command: "/unknown",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Unknown command: /unknown", commandResponse.Text)
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

		plugin, pluginContext := setupPlugin(t, mockAPI, time.Now())

		mockSystem.EXPECT().GetLicense().Return(nil).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				EnableTesting:   &falseValue,
				EnableDeveloper: &falseValue,
			},
		}).Times(1)

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command: "/export",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "The channel export plugin requires a valid Enterprise license.", commandResponse.Text)
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
		plugin, pluginContext := setupPlugin(t, mockAPI, now)

		mockSystem.EXPECT().GetLicense().Return(nil).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				EnableTesting:   &trueValue,
				EnableDeveloper: &trueValue,
			},
		}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(&model.Channel{Id: "channel_id", Name: "channel_name", Type: model.ChannelTypeOpen}, nil)
		mockChannel.EXPECT().GetDirect("user_id", "bot_id").Return(&model.Channel{Id: "direct"}, nil)
		mockUser.EXPECT().HasPermissionTo("user_id", model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})
		mockFile.EXPECT().Upload(gomock.Any(), "channel_name.csv", "direct").Do(func(reader io.Reader, _ /*fileName*/, _ /*channelID*/ string) {
			contents, err := io.ReadAll(reader)
			require.NoError(t, err)
			expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

			require.Equal(t, expected, string(contents))
		}).Return(&model.FileInfo{Id: "file_id"}, nil)
		mockPost.EXPECT().CreatePost(&model.Post{
			UserId:    "bot_id",
			ChannelId: "direct",
			Message:   "Channel ~channel_name exported:",
			FileIds:   []string{"file_id"},
		})

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Exporting ~channel_name. @channelexport will send you a direct message when the export is ready.", commandResponse.Text)

		// Export runs asynchronuosly, so give time for that to occur and complete above
		// mock assertions.
		time.Sleep(1 * time.Second)
	})

	t.Run("not enough permmission to export channel", func(t *testing.T) {
		_, _, _, _, _, mockUser, mockSystem, mockConfiguration, mockCluster, mockAPI := BaseMockSetup(t)
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)
		plugin, pluginContext := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{EnableAdminRestrictions: true})

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockUser.EXPECT().HasPermissionTo("", model.PermissionManageSystem).Return(false)
		mockUser.EXPECT().HasPermissionToChannel("", "channel_id", model.PermissionManageChannelRoles).Return(false)

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "You do not have enough permissions to export this channel", commandResponse.Text)
	})

	t.Run("failed channel fetch", func(t *testing.T) {
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

		plugin, pluginContext := setupPlugin(t, mockAPI, time.Now())

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(nil, errors.New("failed"))
		mockLog.EXPECT().Error("unable to retrieve the channel to export", "Channel ID", "channel_id", "Error", gomock.Any())

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Unable to retrieve the channel to export.", commandResponse.Text)
	})

	t.Run("failed dm channel fetch", func(t *testing.T) {
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

		plugin, pluginContext := setupPlugin(t, mockAPI, time.Now())

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(&model.Channel{Id: "channel_id", Type: model.ChannelTypeOpen}, nil)
		mockChannel.EXPECT().GetDirect("user_id", "bot_id").Return(nil, errors.New("failed"))
		mockLog.EXPECT().Error("unable to create a direct message channel between the bot and the user", "Bot ID", "bot_id", "User ID", "user_id", "Error", gomock.Any())

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "An error occurred trying to create a direct message channel between you and @channelexport.", commandResponse.Text)
	})

	t.Run("export without access to email", func(t *testing.T) {
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
		plugin, pluginContext := setupPlugin(t, mockAPI, now)

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(&model.Channel{Id: "channel_id", Name: "channel_name", Type: model.ChannelTypeOpen}, nil)
		mockChannel.EXPECT().GetDirect("user_id", "bot_id").Return(&model.Channel{Id: "direct"}, nil)
		mockUser.EXPECT().HasPermissionTo("user_id", model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockFile.EXPECT().Upload(gomock.Any(), "channel_name.csv", "direct").Do(func(reader io.Reader, _ /*fileName*/, _ /*channelID*/ string) {
			contents, err := io.ReadAll(reader)
			require.NoError(t, err)
			expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

			require.Equal(t, expected, string(contents))
		}).Return(&model.FileInfo{Id: "file_id"}, nil)
		mockPost.EXPECT().CreatePost(&model.Post{
			UserId:    "bot_id",
			ChannelId: "direct",
			Message:   "Channel ~channel_name exported:",
			FileIds:   []string{"file_id"},
		})

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Exporting ~channel_name. @channelexport will send you a direct message when the export is ready.", commandResponse.Text)

		// Export runs asynchronuosly, so give time for that to occur and complete above
		// mock assertions.
		time.Sleep(1 * time.Second)
	})

	t.Run("export with access to email", func(t *testing.T) {
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
		plugin, pluginContext := setupPlugin(t, mockAPI, now)

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(&model.Channel{Id: "channel_id", Name: "channel_name", Type: model.ChannelTypeOpen}, nil)
		mockChannel.EXPECT().GetDirect("user_id", "bot_id").Return(&model.Channel{Id: "direct"}, nil)
		mockUser.EXPECT().HasPermissionTo("user_id", model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})
		mockFile.EXPECT().Upload(gomock.Any(), "channel_name.csv", "direct").Do(func(reader io.Reader, _ /*fileName*/, _ /*channelID*/ string) {
			contents, err := io.ReadAll(reader)
			require.NoError(t, err)
			expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

			require.Equal(t, expected, string(contents))
		}).Return(&model.FileInfo{Id: "file_id"}, nil)
		mockPost.EXPECT().CreatePost(&model.Post{
			UserId:    "bot_id",
			ChannelId: "direct",
			Message:   "Channel ~channel_name exported:",
			FileIds:   []string{"file_id"},
		})

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Exporting ~channel_name. @channelexport will send you a direct message when the export is ready.", commandResponse.Text)

		// Export runs asynchronuosly, so give time for that to occur and complete above
		// mock assertions.
		time.Sleep(1 * time.Second)
	})

	t.Run("disallow concurrent exports", func(t *testing.T) {
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
		plugin, pluginContext := setupPlugin(t, mockAPI, now)

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).Times(1)
		mockChannel.EXPECT().Get("channel_id").Return(&model.Channel{Id: "channel_id", Name: "channel_name", Type: model.ChannelTypeOpen}, nil)
		mockChannel.EXPECT().GetDirect("user_id", "bot_id").Return(&model.Channel{Id: "direct"}, nil)
		mockUser.EXPECT().HasPermissionTo("user_id", model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})
		wg := sync.WaitGroup{}
		wg.Add(1)
		mockFile.EXPECT().Upload(gomock.Any(), "channel_name.csv", "direct").Do(func(reader io.Reader, _ /*fileName*/, _ /*channelID*/ string) {
			defer wg.Done()
			t.Log("about to sleep for 3s")
			time.Sleep(time.Second * 3)
			_, err := io.ReadAll(reader)
			require.NoError(t, err)
		}).Return(&model.FileInfo{Id: "file_id"}, nil)
		mockPost.EXPECT().CreatePost(&model.Post{
			UserId:    "bot_id",
			ChannelId: "direct",
			Message:   "Channel ~channel_name exported:",
			FileIds:   []string{"file_id"},
		})

		commandResponse, appError := plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "Exporting ~channel_name. @channelexport will send you a direct message when the export is ready.", commandResponse.Text)

		// concurrent executions should fail
		for i := 0; i < 3; i++ {
			commandResponse, appError = plugin.ExecuteCommand(pluginContext, &model.CommandArgs{
				Command:   "/export",
				ChannelId: "channel_id",
				UserId:    "user_id",
			})
			require.Nil(t, appError)
			require.Equal(t, "An export is already running.", commandResponse.Text)
		}

		// wait for upload to complete
		wg.Wait()
	})

	t.Run("no permissions", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		plugin, pluginCtx := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{
			EnableAdminRestrictions: true,
		})

		mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
			FutureFeatures: &trueValue,
		}}).Times(2)
		mockUser.EXPECT().
			HasPermissionToChannel(gomock.Eq("user_id"), gomock.Eq("channel_id"), gomock.Eq(model.PermissionManageChannelRoles)).
			Return(false)
		mockUser.EXPECT().HasPermissionTo("user_id", model.PermissionManageSystem).Return(false).Times(1)
		mockConfiguration.EXPECT().GetConfig().Return(&model.Config{
			PrivacySettings: model.PrivacySettings{
				ShowEmailAddress: &trueValue,
			},
		})

		commandResponse, appError := plugin.ExecuteCommand(pluginCtx, &model.CommandArgs{
			Command:   "/export",
			ChannelId: "channel_id",
			UserId:    "user_id",
		})

		require.Nil(t, appError)
		assert.Equal(t, model.CommandResponseTypeEphemeral, commandResponse.ResponseType)
		assert.Equal(t, "You do not have enough permissions to export this channel", commandResponse.Text)
	})
}

func TestHasPermissionToExportChannel(t *testing.T) {
	t.Run("enable admin restrictions is false", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		plugin, _ := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{
			EnableAdminRestrictions: false,
		})
		require.True(t, plugin.hasPermissionToExportChannel("user_id", "channel_id"))
	})

	t.Run("user has permission to manage channel roles", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		exporterMutex := pluginapi.NewClusterMutexMock()
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(exporterMutex, nil)

		plugin, _ := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{
			EnableAdminRestrictions: true,
		})
		mockUser.EXPECT().
			HasPermissionToChannel(gomock.Eq("user_id"), gomock.Eq("channel_id"), gomock.Eq(model.PermissionManageChannelRoles)).
			Return(true)

		hasPermission := plugin.hasPermissionToExportChannel("user_id", "channel_id")
		assert.True(t, hasPermission)
	})

	t.Run("user has permission to manage system", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		exporterMutex := pluginapi.NewClusterMutexMock()
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(exporterMutex, nil)

		plugin, _ := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{
			EnableAdminRestrictions: true,
		})
		mockUser.EXPECT().
			HasPermissionToChannel(gomock.Eq("user_id"), gomock.Eq("channel_id"), gomock.Eq(model.PermissionManageChannelRoles)).
			Return(false)

		mockUser.EXPECT().
			HasPermissionTo(gomock.Eq("user_id"), gomock.Eq(model.PermissionManageSystem)).
			Return(true)

		hasPermission := plugin.hasPermissionToExportChannel("user_id", "channel_id")
		assert.True(t, hasPermission)
	})

	t.Run("user don't have required permissions", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

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

		exporterMutex := pluginapi.NewClusterMutexMock()
		mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(exporterMutex, nil)

		plugin, _ := setupPlugin(t, mockAPI, time.Now())
		plugin.setConfiguration(&configuration{
			EnableAdminRestrictions: true,
		})
		mockUser.EXPECT().
			HasPermissionToChannel(gomock.Eq("user_id"), gomock.Eq("channel_id"), gomock.Eq(model.PermissionManageChannelRoles)).
			Return(false)

		mockUser.EXPECT().
			HasPermissionTo(gomock.Eq("user_id"), gomock.Eq(model.PermissionManageSystem)).
			Return(false)

		hasPermission := plugin.hasPermissionToExportChannel("user_id", "channel_id")
		assert.False(t, hasPermission)
	})
}

type MockUser struct {
	mock.Mock
}

func (m *MockUser) HasPermissionToChannel(userID, channelID string, permission *model.Permission) bool {
	args := m.Called(userID, channelID, permission)
	return args.Bool(0)
}

func (m *MockUser) HasPermissionTo(userID string, permission *model.Permission) bool {
	args := m.Called(userID, permission)
	return args.Bool(0)
}

func (m *MockUser) Get(userID string) (*model.User, error) {
	args := m.Called(userID)
	return args.Get(0).(*model.User), args.Error(1)
}

func TestPermissionToExportChannel(t *testing.T) {
	mockUser := new(MockUser)
	clientWrapper := &pluginapi.Wrapper{
		User: mockUser,
	}
	plugin := &Plugin{
		client: clientWrapper,
	}

	tests := []struct {
		name                    string
		enableAdminRestrictions bool
		userID                  string
		channelID               string
		hasPermissionToChannel  bool
		hasPermissionToSystem   bool
		expected                bool
	}{
		{
			name:                    "Admin restrictions enabled and user has permission to channel",
			enableAdminRestrictions: true,
			userID:                  "mockUser1",
			channelID:               "mockChannel1",
			hasPermissionToChannel:  true,
			hasPermissionToSystem:   false,
			expected:                true,
		},
		{
			name:                    "Admin restrictions enabled and user has permission to system",
			enableAdminRestrictions: true,
			userID:                  "mockUser2",
			channelID:               "mockChannel2",
			hasPermissionToChannel:  false,
			hasPermissionToSystem:   true,
			expected:                true,
		},
		{
			name:                    "Admin restrictions enabled and user lacks both system and channel permissions",
			enableAdminRestrictions: true,
			userID:                  "mockUser3",
			channelID:               "mockChannel3",
			hasPermissionToChannel:  false,
			hasPermissionToSystem:   false,
			expected:                false,
		},
		{
			name:                    "Admin restrictions disabled",
			enableAdminRestrictions: false,
			userID:                  "mockUser4",
			channelID:               "mockChannel4",
			hasPermissionToChannel:  false,
			hasPermissionToSystem:   false,
			expected:                true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.enableAdminRestrictions {
				mockUser.On("HasPermissionToChannel", tt.userID, tt.channelID, model.PermissionManageChannelRoles).Return(tt.hasPermissionToChannel)
				if !tt.hasPermissionToChannel {
					mockUser.On("HasPermissionTo", tt.userID, model.PermissionManageSystem).Return(tt.hasPermissionToSystem)
				}
			}

			plugin.configuration = &configuration{
				EnableAdminRestrictions: tt.enableAdminRestrictions,
			}

			resp := plugin.hasPermissionToExportChannel(tt.userID, tt.channelID)
			assert.Equal(t, tt.expected, resp)
			mockUser.AssertExpectations(t)
		})
	}
}

func TestUploadFileTo(t *testing.T) {
	trueValue := true

	testCases := []struct {
		name          string
		setupMocks    func(mockFile *mock_pluginapi.MockFile, mockLog *mock_pluginapi.MockLog, mockSystem *mock_pluginapi.MockSystem, mockConfiguration *mock_pluginapi.MockConfiguration)
		expectedError string
		expectedFile  *model.FileInfo
	}{
		{
			name: "file upload error",
			setupMocks: func(mockFile *mock_pluginapi.MockFile, mockLog *mock_pluginapi.MockLog, mockSystem *mock_pluginapi.MockSystem, mockConfiguration *mock_pluginapi.MockConfiguration) {
				mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
					FutureFeatures: &trueValue,
				}}).AnyTimes()
				mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).AnyTimes()
				mockLog.EXPECT().Error("unable to upload the exported file to the channel", "Channel ID", "mockChannelID", "Error", gomock.Any())
				mockFile.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("file upload error"))
			},
			expectedError: "unable to upload the exported file: file upload error",
		},
		{
			name: "file upload successful",
			setupMocks: func(mockFile *mock_pluginapi.MockFile, _ *mock_pluginapi.MockLog, mockSystem *mock_pluginapi.MockSystem, mockConfiguration *mock_pluginapi.MockConfiguration) {
				mockSystem.EXPECT().GetLicense().Return(&model.License{Features: &model.Features{
					FutureFeatures: &trueValue,
				}}).AnyTimes()
				mockConfiguration.EXPECT().GetConfig().Return(&model.Config{}).AnyTimes()
				var file *model.FileInfo
				mockFile.EXPECT().Upload(gomock.Any(), gomock.Any(), gomock.Any()).Return(file, nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, mockFile, mockLog, _, _, _, mockSystem, mockConfiguration, mockCluster, mockAPI := BaseMockSetup(t)

			mockCluster.EXPECT().NewMutex(gomock.Eq(KeyClusterMutex)).Return(pluginapi.NewClusterMutexMock(), nil)
			tt.setupMocks(mockFile, mockLog, mockSystem, mockConfiguration)

			plugin, _ := setupPlugin(t, mockAPI, time.Now())
			plugin.setConfiguration(&configuration{EnableAdminRestrictions: true})

			var content io.Reader
			uploadedFile, err := plugin.uploadFileTo("mockFileName", content, "mockChannelID")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFile, uploadedFile)
			}
		})
	}
}

func BaseMockSetup(t *testing.T) (
	*mock_pluginapi.MockChannel,
	*mock_pluginapi.MockFile,
	*mock_pluginapi.MockLog,
	*mock_pluginapi.MockPost,
	*mock_pluginapi.MockSlashCommand,
	*mock_pluginapi.MockUser,
	*mock_pluginapi.MockSystem,
	*mock_pluginapi.MockConfiguration,
	*mock_pluginapi.MockCluster,
	*pluginapi.Wrapper,
) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

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

	return mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration, mockCluster, mockAPI
}
