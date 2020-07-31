package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
)

func makeTestPostsIterator(t *testing.T, now time.Time) func(channel *model.Channel, showEmailAddress bool) PostIterator {
	exportedPosts := []*ExportedPost{
		{
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

	return func(channel *model.Channel, showEmailAddress bool) PostIterator {
		return func() ([]*ExportedPost, error) {
			retExportedPosts := exportedPosts
			if !showEmailAddress {
				for i := range retExportedPosts {
					retExportedPosts[i].UserEmail = ""
				}
			}

			// Once consumed, mark it as nil so the iterator ends.
			exportedPosts = nil

			return retExportedPosts, nil
		}
	}
}
