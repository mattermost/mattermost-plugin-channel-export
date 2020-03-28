package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type CSVExporter struct {
}

func (e *CSVExporter) Export(nextPosts PostIterator, writer io.Writer) error {
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
		return fmt.Errorf("Unable to create a CSV file: %w", err)
	}

	batchSize := 100
	for {
		posts, err := nextPosts(batchSize)
		if err != nil {
			return fmt.Errorf("unable to retrieve next %d posts: %w", batchSize, err)
		}

		for _, post := range posts {
			csvWriter.Write([]string{
				strconv.FormatInt(post.CreateAt, 10),
				post.UserID,
				post.UserEmail,
				post.UserType,
				post.UserName,
				post.ID,
				post.ParentPostID,
				post.Message,
				post.Type,
			})
		}

		if len(posts) < batchSize {
			break
		}
	}

	csvWriter.Flush()

	return nil
}
