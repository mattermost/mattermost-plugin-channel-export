package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"
)

func setupAPI(t *testing.T, mockAPI *pluginapi.Wrapper, now time.Time, userID, channelID string) string {
	exportedPosts := []*ExportedPost{
		&ExportedPost{
			CreateAt:     now.Round(time.Millisecond).UTC(),
			UserID:       "post_user_id",
			UserEmail:    "post_user_email",
			UserType:     "user",
			UserName:     "post_user_nickname",
			ID:           "post_id",
			ParentPostID: "post_parent_id",
			Message:      "post_message",
			Type:         "message",
		},
	}

	router := mux.NewRouter()
	registerAPI(router, mockAPI, func(channel *model.Channel) PostIterator {
		return func() ([]*ExportedPost, error) {
			retExportedPosts := exportedPosts
			exportedPosts = nil

			return retExportedPosts, nil
		}
	})

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
	t.Run("unauthorized", func(t *testing.T) {
		address := setupAPI(t, nil, time.Now(), "", "channel_id")
		client := NewClient(address)

		err := client.ExportChannel(ioutil.Discard, "channel_id", FormatCSV)
		require.EqualError(t, err, "failed with status code 401")
	})

	t.Run("missing channel_id", func(t *testing.T) {
		address := setupAPI(t, nil, time.Now(), "user_id", "channel_id")
		client := NewClient(address)
		client.SetToken("token")

		err := client.ExportChannel(ioutil.Discard, "", FormatCSV)
		require.EqualError(t, err, "missing channel_id parameter")
	})

	t.Run("missing format", func(t *testing.T) {
		address := setupAPI(t, nil, time.Now(), "user_id", "channel_id")
		client := NewClient(address)
		client.SetToken("token")

		err := client.ExportChannel(ioutil.Discard, "channel_id", "")
		require.EqualError(t, err, "missing format parameter")
	})

	t.Run("unsupported format", func(t *testing.T) {
		address := setupAPI(t, nil, time.Now(), "user_id", "channel_id")
		client := NewClient(address)
		client.SetToken("token")

		err := client.ExportChannel(ioutil.Discard, "channel_id", "pdf2")
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

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser)

		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), "user_id", channelID)
		client := NewClient(address)
		client.SetToken("token")

		mockChannel.EXPECT().Get(channelID).Return(nil, &model.AppError{StatusCode: http.StatusNotFound}).Times(1)

		err := client.ExportChannel(ioutil.Discard, channelID, FormatCSV)
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

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser)

		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), "user_id", channelID)
		client := NewClient(address)
		client.SetToken("token")

		mockChannel.EXPECT().Get(channelID).Return(nil, &model.AppError{StatusCode: http.StatusInternalServerError}).Times(1)

		err := client.ExportChannel(ioutil.Discard, channelID, FormatCSV)
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

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser)

		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, time.Now(), userID, channelID)
		client := NewClient(address)
		client.SetToken("token")

		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PERMISSION_READ_CHANNEL).Return(false).Times(1)

		err := client.ExportChannel(ioutil.Discard, channelID, FormatCSV)
		require.EqualError(t, err, "channel 'channel_id' not found or user does not have permission")
	})

	t.Run("export with channel read permission", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)

		mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
		mockFile := mock_pluginapi.NewMockFile(mockCtrl)
		mockLog := mock_pluginapi.NewMockLog(mockCtrl)
		mockPost := mock_pluginapi.NewMockPost(mockCtrl)
		mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
		mockUser := mock_pluginapi.NewMockUser(mockCtrl)

		mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser)

		now := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.FixedZone("UTC-8", -8*60*60))
		userID := "user_id"
		channelID := "channel_id"
		address := setupAPI(t, mockAPI, now, userID, channelID)
		client := NewClient(address)
		client.SetToken("token")

		mockChannel.EXPECT().Get(channelID).Return(&model.Channel{Id: channelID}, nil).Times(1)
		mockUser.EXPECT().HasPermissionToChannel(userID, channelID, model.PERMISSION_READ_CHANNEL).Return(true).Times(1)

		var buffer bytes.Buffer
		err := client.ExportChannel(&buffer, channelID, FormatCSV)
		require.NoError(t, err)

		expected := `Post Creation Time,User Id,User Email,User Type,User Name,Post Id,Parent Post Id,Post Message,Post Type
2009-11-11 07:00:00 +0000 UTC,post_user_id,post_user_email,user,post_user_nickname,post_id,post_parent_id,post_message,message
`

		require.Equal(t, expected, string(buffer.Bytes()))
	})
}
