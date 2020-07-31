package main

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CSV exports all the posts in a channel to a chronollogically
// ordered file in CSV format
type CSV struct{}

// FileName returns the passed name with the .csv extension added
func (e *CSV) FileName(name string) string {
	return fmt.Sprintf("%s.csv", name)
}

// ContentType returns the content type of the file format being exported.
func (e *CSV) ContentType() string {
	return "text/csv"
}

// Export consumes all the posts returned by the iterator and writes them in
// CSV format to the writer
func (e *CSV) Export(nextPosts PostIterator, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	err := csvWriter.Write([]string{
		"Post Creation Time",
		"User Id",
		"User Email",
		"User Type",
		"User Name",
		"Post Id",
		"Parent Post Id",
		"Post Message",
		"Post Type",
	})
	if err != nil {
		return errors.Wrap(err, "unable to create a CSV file")
	}

	for {
		posts, err := nextPosts()
		if err != nil {
			return errors.Wrap(err, "unable to retrieve next posts")
		}

		for _, post := range posts {
			err = csvWriter.Write([]string{
				post.CreateAt.String(),
				post.UserID,
				post.UserEmail,
				post.UserType,
				post.UserName,
				post.ID,
				post.ParentPostID,
				post.Message,
				post.Type,
			})
			if err != nil {
				return errors.Wrap(err, "unable to write csv")
			}
		}

		if len(posts) == 0 {
			break
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		logrus.WithError(err).Warn("failed to flush CSV file")
	}

	return nil
}
