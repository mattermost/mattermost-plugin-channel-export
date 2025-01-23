// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package util

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestLimitPipeWriter_WriteWithinLimit(t *testing.T) {
	r, w := io.Pipe()
	lpw := NewLimitPipeWriter(w, 10) // 10 bytes limit
	data := []byte("hello")

	go func() {
		defer w.Close()
		_, err := lpw.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Errorf("unexpected error during io.Copy: %v", err)
	}

	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got %s", buf.String())
	}
}

func TestLimitPipeWriter_WriteExceedsLimit(t *testing.T) {
	r, w := io.Pipe()
	lpw := NewLimitPipeWriter(w, 5) // 5 bytes limit
	data := []byte("hello world")

	go func() {
		defer w.Close()
		_, err := lpw.Write(data)
		if err == nil {
			t.Error("expected an error due to limit exceeded but got none")
		} else if err.Error() != "limit (5 bytes) exceeded" {
			t.Errorf("unexpected error message: %v", err)
		}
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	// Expect an error due to exceeding limit
	if err == nil {
		t.Error("expected an error during io.Copy due to limit exceeded, but got none")
	}

	if buf.String() != "" {
		t.Errorf("expected empty buffer due to limit exceeded, got %s", buf.String())
	}
}

func TestLimitPipeWriter_MultipleWrites(t *testing.T) {
	r, w := io.Pipe()
	lpw := NewLimitPipeWriter(w, 10) // 10 bytes limit

	go func() {
		defer w.Close()
		data1 := []byte("hello")
		data2 := []byte("world")
		_, err := lpw.Write(data1)
		if err != nil {
			t.Errorf("unexpected error writing first data block: %v", err)
		}
		_, err = lpw.Write(data2)
		if err != nil {
			t.Errorf("unexpected error writing second data block: %v", err)
		}
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Errorf("unexpected error during io.Copy: %v", err)
	}

	if buf.String() != "helloworld" {
		t.Errorf("expected 'helloworld', got %s", buf.String())
	}
}

func TestLimitPipeWriter_WriteUpToLimit(t *testing.T) {
	r, w := io.Pipe()
	lpw := NewLimitPipeWriter(w, 5) // 5 bytes limit

	go func() {
		defer w.Close()
		data := []byte("hello") // exactly 5 bytes
		_, err := lpw.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		t.Errorf("unexpected error during io.Copy: %v", err)
	}

	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got %s", buf.String())
	}
}

func TestLimitPipeWriter_CloseWithError(t *testing.T) {
	r, w := io.Pipe()
	lpw := NewLimitPipeWriter(w, 10)
	expectedErr := fmt.Errorf("oh no")

	go func() {
		data := []byte("hello")
		_, err := lpw.Write(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Close the writer with an error
		err = lpw.CloseWithError(expectedErr)
		if err != nil {
			t.Errorf("unexpected error closing pipe with error: %v", err)
		}
	}()

	// Attempt to read, expecting a pipe error after first read
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if buf.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", buf.String())
	}
}
