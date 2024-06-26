package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
	"github.com/mattermost/mattermost-server/v6/model"
)

func TestChannelPostsIterator(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
	mockFile := mock_pluginapi.NewMockFile(mockCtrl)
	mockLog := mock_pluginapi.NewMockLog(mockCtrl)
	mockPost := mock_pluginapi.NewMockPost(mockCtrl)
	mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
	mockUser := mock_pluginapi.NewMockUser(mockCtrl)
	mockSystem := mock_pluginapi.NewMockSystem(mockCtrl)
	mockConfiguration := mock_pluginapi.NewMockConfiguration(mockCtrl)

	mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration)

	channel := &model.Channel{
		Id: "jx2289hnvko3dypmc3thfcafpb",
	}

	now := time.Now().Round(time.Millisecond)
	userID := "h6itnszvtit5k2jhi2c1o3p7ox"

	user := model.User{
		Id:       userID,
		Email:    "alex@example.com",
		Nickname: "alex",
		IsBot:    false,
	}

	post := model.Post{
		UserId:   userID,
		Type:     "",
		CreateAt: model.GetMillisForTime(now),
		Id:       "3j6wc01x7ox5joy3jupjmo69zu",
		Message:  "test",
	}

	postList := model.PostList{
		Posts: map[string]*model.Post{
			post.Id: &post,
		},
		Order: []string{post.Id},
	}

	t.Run("One post iterator", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel, false)

		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, PerPage).Return(&postList, nil).Times(1)
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		posts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, posts, 1)
	})

	t.Run("Paging is correct", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel, false)

		length := PerPage
		posts := make(map[string]*model.Post, length)
		order := make([]string, length)

		for i := 0; i < length; i++ {
			id := strconv.Itoa(i)
			var newPost model.Post
			require.NoError(t, post.ShallowCopy(&newPost))
			newPost.Id = id

			order[i] = id
			posts[id] = &newPost
		}

		firstPage := model.PostList{
			Posts: posts,
			Order: order,
		}

		secondPage := postList

		gomock.InOrder(
			mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, PerPage).Return(&firstPage, nil).Times(1),
			mockPost.EXPECT().GetPostsForChannel(channel.Id, 1, PerPage).Return(&secondPage, nil).Times(1),
		)

		// Called only once because we are setting the same user in all posts,
		// so there is always a cache hit
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		firstPagePosts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, firstPagePosts, length)

		for i, post := range firstPagePosts {
			require.Equal(t, strconv.Itoa(i), post.ID)
		}

		secondPagePosts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, secondPagePosts, 1)
	})

	t.Run("Old posts with a new version are skipped", func(t *testing.T) {
		var editedPost model.Post
		require.NoError(t, post.ShallowCopy(&editedPost))
		editedPost.OriginalId = "original_id"

		postIterator := channelPostsIterator(mockAPI, channel, false)

		editedPostList := model.PostList{
			Posts: map[string]*model.Post{
				post.Id: &editedPost,
			},
			Order: []string{editedPost.Id},
		}

		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, PerPage).Return(&editedPostList, nil).Times(1)

		posts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, posts, 0)
	})

	t.Run("Error when retreiving posts is moved up to the caller", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel, false)

		expectedError := errors.New("error retreiving posts")
		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, PerPage).Return(nil, expectedError).Times(1)

		posts, err := postIterator()
		require.Nil(t, posts)
		require.Equal(t, expectedError, err)
	})

	t.Run("Error when exporting a post is moved up to the caller", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel, false)

		expectedError := fmt.Errorf("new error")
		mockUser.EXPECT().Get(post.UserId).Return(nil, expectedError).Times(1)
		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, PerPage).Return(&postList, nil).Times(1)

		posts, err := postIterator()
		require.Nil(t, posts)
		require.True(t, errors.Is(err, expectedError))
	})
}

func TestToExportedPost(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	mockChannel := mock_pluginapi.NewMockChannel(mockCtrl)
	mockFile := mock_pluginapi.NewMockFile(mockCtrl)
	mockLog := mock_pluginapi.NewMockLog(mockCtrl)
	mockPost := mock_pluginapi.NewMockPost(mockCtrl)
	mockSlashCommand := mock_pluginapi.NewMockSlashCommand(mockCtrl)
	mockUser := mock_pluginapi.NewMockUser(mockCtrl)
	mockSystem := mock_pluginapi.NewMockSystem(mockCtrl)
	mockConfiguration := mock_pluginapi.NewMockConfiguration(mockCtrl)

	mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem, mockConfiguration)

	now := time.Now().Round(time.Millisecond)
	userID := "h6itnszvtit5k2jhi2c1o3p7ox"

	post := model.Post{
		UserId:   userID,
		Type:     "",
		CreateAt: model.GetMillisForTime(now),
		Id:       "3j6wc01x7ox5joy3jupjmo69zu",
		Message:  "test",
	}

	user := model.User{
		Id:       userID,
		Username: "alex",
		Email:    "alex@example.com",
		Nickname: "alex is the man",
		IsBot:    false,
	}

	t.Run("Normal message, don't show email address", func(t *testing.T) {
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &post, false, usersCache)
		require.NoError(t, err)

		exportedPost := ExportedPost{
			CreateAt:  now.UTC(),
			UserID:    post.UserId,
			UserEmail: "",
			UserType:  "user",
			UserName:  user.Username,
			ID:        post.Id,
			Message:   post.Message,
			Type:      "message",
		}
		require.Equal(t, &exportedPost, actualExportedPost)
	})

	t.Run("Normal message, show email address", func(t *testing.T) {
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &post, true, usersCache)
		require.NoError(t, err)

		exportedPost := ExportedPost{
			CreateAt:  now.UTC(),
			UserID:    post.UserId,
			UserEmail: user.Email,
			UserType:  "user",
			UserName:  user.Username,
			ID:        post.Id,
			Message:   post.Message,
			Type:      "message",
		}
		require.Equal(t, &exportedPost, actualExportedPost)
	})

	t.Run("User not found", func(t *testing.T) {
		var postWithoutUserID model.Post
		require.NoError(t, post.ShallowCopy(&postWithoutUserID))
		postWithoutUserID.UserId = "unknown_user_id"

		err := fmt.Errorf("new error")
		mockUser.EXPECT().Get(postWithoutUserID.UserId).Return(nil, err).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &postWithoutUserID, false, usersCache)
		require.Error(t, err)
		require.Nil(t, actualExportedPost)
	})

	t.Run("Bot message", func(t *testing.T) {
		bot := user
		bot.IsBot = true

		mockUser.EXPECT().Get(post.UserId).Return(&bot, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &post, false, usersCache)
		require.NoError(t, err)

		expectedPost := ExportedPost{
			CreateAt:  now.UTC(),
			UserID:    post.UserId,
			UserEmail: "",
			UserType:  "bot",
			UserName:  user.Username,
			ID:        post.Id,
			Message:   post.Message,
			Type:      "message",
		}
		require.Equal(t, &expectedPost, actualExportedPost)
	})

	t.Run("System message", func(t *testing.T) {
		var systemPost model.Post
		require.NoError(t, post.ShallowCopy(&systemPost))
		systemPost.Type = "system_join_channel"

		mockUser.EXPECT().Get(systemPost.UserId).Return(&user, nil).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &systemPost, false, usersCache)
		require.NoError(t, err)

		expectedPost := ExportedPost{
			CreateAt:  now.UTC(),
			UserID:    post.UserId,
			UserEmail: "",
			UserType:  "system",
			UserName:  user.Username,
			ID:        post.Id,
			Message:   post.Message,
			Type:      systemPost.Type,
		}

		require.Equal(t, &expectedPost, actualExportedPost)
	})
}
