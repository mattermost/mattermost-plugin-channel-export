package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileName(t *testing.T) {
	exporter := CSVExporter{}

	testCases := []struct {
		testName         string
		name             string
		expectedFilename string
	}{
		{"Empty name", "", ".csv"},
		{"Normal name", "name", "name.csv"},
		{"Name with unicode chars", "αβ", "αβ.csv"},
		{"Name with digits", "1", "1.csv"},
	}

	for _, test := range testCases {
		t.Run(test.testName, func(*testing.T) {
			require.Equal(t, exporter.FileName(test.name), test.expectedFilename)
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

var exportedPost = &ExportedPost{
	CreateAt:     1,
	UserID:       "dummyUserID",
	UserEmail:    "dummy@email.com",
	UserType:     "user",
	UserName:     "dummy",
	ID:           "dummyPostID",
	ParentPostID: "",
	Message:      "Lorem ipsum",
	Type:         "message",
}

func exportedPostToCSV(post *ExportedPost) string {
	fields := []string{
		strconv.FormatInt(post.CreateAt, 10),
		post.UserID,
		post.UserEmail,
		post.UserType,
		post.UserName,
		post.ID,
		post.ParentPostID,
		post.Message,
		post.Type,
	}
	return strings.Join(fields, ",") + "\n"
}

func TestExport(t *testing.T) {
	header := []string{
		"Post Creation Time",
		"User Id",
		"User Email",
		"User Type",
		"User Name",
		"Post Id",
		"Parent Post Id",
		"Post Message",
		"Post Type",
	}
	headerCSV := strings.Join(header, ",") + "\n"

	genIterator := func(numPosts, batchSize int) PostIterator {
		sent := 0
		return func() ([]*ExportedPost, error) {
			if sent >= numPosts {
				return nil, nil
			}

			length := min(numPosts-sent, batchSize)

			posts := make([]*ExportedPost, length)
			for i := 0; i < length; i++ {
				posts[i] = exportedPost
			}

			sent += length
			return posts, nil
		}
	}

	exporter := CSVExporter{}

	t.Run("Empty iterator", func(t *testing.T) {
		var actualString strings.Builder

		err := exporter.Export(genIterator(0, 0), &actualString)

		require.Nil(t, err)
		require.Equal(t, headerCSV, actualString.String())
	})

	t.Run("One post", func(t *testing.T) {
		var actualString strings.Builder

		err := exporter.Export(genIterator(1, 1), &actualString)

		require.Nil(t, err)
		require.Equal(t, headerCSV+exportedPostToCSV(exportedPost), actualString.String())
	})

	t.Run("Several posts", func(t *testing.T) {
		var actualString strings.Builder

		err := exporter.Export(genIterator(10, 4), &actualString)

		expected := headerCSV
		for i := 0; i < 10; i++ {
			expected += exportedPostToCSV(exportedPost)
		}

		require.Nil(t, err)
		require.Equal(t, expected, actualString.String())
	})

	t.Run("Wrong iterator", func(t *testing.T) {
		var actualString strings.Builder

		err := exporter.Export(
			func() ([]*ExportedPost, error) { return nil, errors.New("forcing an error") },
			&actualString,
		)

		require.Error(t, err)
	})
}
