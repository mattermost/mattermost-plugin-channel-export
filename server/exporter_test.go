package main

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi/mock_pluginapi"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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

	mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem)

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
		ParentId: "o3p7oxj1yqtnwg66u95802y08j",
		Message:  "test",
	}

	postList := model.PostList{
		Posts: map[string]*model.Post{
			post.Id: &post,
		},
		Order: []string{post.Id},
	}

	t.Run("One post iterator", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel)

		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, 1000).Return(&postList, nil).Times(1)
		mockUser.EXPECT().Get(post.UserId).Return(&user, nil).Times(1)

		posts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, posts, 1)
	})

	t.Run("Paging is correct", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel)

		length := 1000
		posts := make(map[string]*model.Post, length)
		order := make([]string, length)

		for i := 0; i < length; i++ {
			id := strconv.Itoa(i)
			var newPost model.Post
			post.ShallowCopy(&newPost)
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
			mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, 1000).Return(&firstPage, nil).Times(1),
			mockPost.EXPECT().GetPostsForChannel(channel.Id, 1, 1000).Return(&secondPage, nil).Times(1),
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
		post.ShallowCopy(&editedPost)
		editedPost.OriginalId = "original_id"

		postIterator := channelPostsIterator(mockAPI, channel)

		editedPostList := model.PostList{
			Posts: map[string]*model.Post{
				post.Id: &editedPost,
			},
			Order: []string{editedPost.Id},
		}

		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, 1000).Return(&editedPostList, nil).Times(1)

		posts, err := postIterator()
		require.NoError(t, err)
		require.Len(t, posts, 0)
	})

	t.Run("Error when retreiving posts is moved up to the caller", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel)

		expectedError := errors.New("error retreiving posts")
		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, 1000).Return(nil, expectedError).Times(1)

		posts, err := postIterator()
		require.Nil(t, posts)
		require.Equal(t, expectedError, err)
	})

	t.Run("Error when exporting a post is moved up to the caller", func(t *testing.T) {
		postIterator := channelPostsIterator(mockAPI, channel)

		expectedError := fmt.Errorf("new error")
		mockUser.EXPECT().Get(post.UserId).Return(nil, expectedError).Times(1)
		mockPost.EXPECT().GetPostsForChannel(channel.Id, 0, 1000).Return(&postList, nil).Times(1)

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

	mockAPI := pluginapi.CustomWrapper(mockChannel, mockFile, mockLog, mockPost, mockSlashCommand, mockUser, mockSystem)

	now := time.Now().Round(time.Millisecond)
	userID := "h6itnszvtit5k2jhi2c1o3p7ox"

	post := model.Post{
		UserId:   userID,
		Type:     "",
		CreateAt: model.GetMillisForTime(now),
		Id:       "3j6wc01x7ox5joy3jupjmo69zu",
		ParentId: "o3p7oxj1yqtnwg66u95802y08j",
		Message:  "test",
	}

	user := model.User{
		Id:       userID,
		Email:    "alex@example.com",
		Nickname: "alex",
		IsBot:    false,
	}

	exportedPost := ExportedPost{
		CreateAt:     now.UTC(),
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
		var postWithoutUserID model.Post
		post.ShallowCopy(&postWithoutUserID)
		postWithoutUserID.UserId = "unknown_user_id"

		error := fmt.Errorf("new error")
		mockUser.EXPECT().Get(postWithoutUserID.UserId).Return(nil, error).Times(1)

		usersCache := make(map[string]*model.User)
		actualExportedPost, err := toExportedPost(mockAPI, &postWithoutUserID, usersCache)
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
		var systemPost model.Post
		post.ShallowCopy(&systemPost)
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
