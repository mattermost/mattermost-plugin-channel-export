package main

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-channel-export/server/apiwrapper"
	"github.com/mattermost/mattermost-plugin-channel-export/server/apiwrapper/mock_apiwrapper"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"
)

func TestToExportedPost(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockChannel := mock_apiwrapper.NewMockChannel(mockCtrl)
	mockFile := mock_apiwrapper.NewMockFile(mockCtrl)
	mockLog := mock_apiwrapper.NewMockLog(mockCtrl)
	mockPost := mock_apiwrapper.NewMockPost(mockCtrl)
	mockSlashCommand := mock_apiwrapper.NewMockSlashCommand(mockCtrl)
	mockUser := mock_apiwrapper.NewMockUser(mockCtrl)

	mockAPI := apiwrapper.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser)

	var timestamp int64 = 1586520040073
	userId := "h6itnszvtit5k2jhi2c1o3p7ox"

	post := model.Post{
		UserId:   userId,
		Type:     "",
		CreateAt: timestamp,
		Id:       "3j6wc01x7ox5joy3jupjmo69zu",
		ParentId: "o3p7oxj1yqtnwg66u95802y08j",
		Message:  "test",
	}

	user := model.User{
		Id:       userId,
		Email:    "alex@example.com",
		Nickname: "alex",
		IsBot:    false,
	}

	exportedPost := ExportedPost{
		CreateAt:     millisToUnix(post.CreateAt),
		UserID:       post.UserId,
		UserEmail:    user.Email,
		UserType:     "user",
		UserName:     user.Nickname,
		ID:           post.Id,
		ParentPostID: post.ParentId,
		Message:      post.Message,
		Type:         "message",
	}

	t.Run("Normal message", func(t *testing.T) {
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &post, usersCache)
		require.NoError(t, err)

		require.Equal(t, &exportedPost, actualExportedPost)
	})

	t.Run("User not found", func(t *testing.T) {
		postWithoutUserId := post
		postWithoutUserId.UserId = "unknown_user_id"

		error := fmt.Errorf("new error")
		mockUser.EXPECT().Get(postWithoutUserId.UserId).Return(nil, error).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &postWithoutUserId, usersCache)
		require.Error(t, err)
		require.Nil(t, actualExportedPost)
	})

	t.Run("Bot message", func(t *testing.T) {
		bot := user
		bot.IsBot = true

		mockUser.EXPECT().Get(post.UserId).Return(&bot, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &post, usersCache)
		require.NoError(t, err)

		expectedPost := exportedPost
		expectedPost.UserType = "bot"

		require.Equal(t, &expectedPost, actualExportedPost)
	})

	t.Run("System message", func(t *testing.T) {
		systemPost := post
		systemPost.Type = "system_join_channel"

		mockUser.EXPECT().Get(systemPost.UserId).Return(&user, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &systemPost, usersCache)
		require.NoError(t, err)

		expectedPost := exportedPost
		expectedPost.UserType = "system"
		expectedPost.Type = systemPost.Type

		require.Equal(t, &expectedPost, actualExportedPost)
	})
}
