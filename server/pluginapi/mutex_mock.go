package pluginapi

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

var (
	ErrLockTimeout = errors.New("timeout")
)

type ClusterMutexMock struct {
	locked int32
}

func NewClusterMutexMock() *ClusterMutexMock {
	return &ClusterMutexMock{}
}

func (m *ClusterMutexMock) LockWithContext(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if atomic.CompareAndSwapInt32(&m.locked, 0, 1) {
				// we have the lock
				return nil
			}
		}
		time.Sleep(time.Millisecond * 20)
	}
}

func (m *ClusterMutexMock) Unlock() {
	if !atomic.CompareAndSwapInt32(&m.locked, 1, 0) {
		panic("not locked")
	}
}
