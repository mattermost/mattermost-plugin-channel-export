// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package util

import (
	"fmt"
	"io"
	"sync"
)

type ErrLimitExceeded struct {
	Limit uint64
}

func (e ErrLimitExceeded) Error() string {
	return fmt.Sprintf("limit (%d bytes) exceeded", e.Limit)
}

// LimitPipeWriter wraps an io.PipeWriter and provides a limit to the number of bytes that can be written.
// Exceeding the limit will cause the current Write call to error, and the underlying PipeWriter will be
// closed, causing subsequent reads on the pipe to error.
type LimitPipeWriter struct {
	pw    *io.PipeWriter
	limit uint64
	count uint64
	mux   sync.Mutex // protects limit, count
}

// NewLimitPipeWriter creates a new LimitPipeWriter with the specified limit.
func NewLimitPipeWriter(pw *io.PipeWriter, limit uint64) *LimitPipeWriter {
	return &LimitPipeWriter{
		pw:    pw,
		limit: limit,
	}
}

// Write implements io.Writer
func (lpw *LimitPipeWriter) Write(p []byte) (int, error) {
	lpw.mux.Lock()
	defer lpw.mux.Unlock()

	count := uint64(len(p))
	if lpw.count+count > lpw.limit {
		err := ErrLimitExceeded{lpw.limit}
		lpw.pw.CloseWithError(err) // ok to call multiple times
		return 0, err
	}

	n, err := lpw.pw.Write(p)
	// According to io.Writer interface contract, n should always be >= 0
	// However, we still check to satisfy the gosec G115 warning about int to uint64 conversion
	if n >= 0 {
		lpw.count += uint64(n)
	}
	return n, err
}

// CloseWithError closes the writer. Future reads from the underlying pipe will return the specified error.
// It is safe to call this multiple times - subsequent calls to CloseWithError will be a no-op.
func (lpw *LimitPipeWriter) CloseWithError(err error) error {
	return lpw.pw.CloseWithError(err)
}

// Close closes the writer; subsequent reads from the
// read half of the pipe will return no bytes and io.EOF.
func (lpw *LimitPipeWriter) Close() error {
	return lpw.pw.Close()
}
