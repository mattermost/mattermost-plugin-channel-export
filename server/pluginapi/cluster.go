package pluginapi

import (
	"context"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type ClusterMutex interface {
	LockWithContext(ctx context.Context) error
	Unlock()
}

// ClusterService exposes methods from the mm server cluster package.
type ClusterService struct {
	api plugin.API
}

func NewClusterService(api plugin.API) *ClusterService {
	return &ClusterService{
		api: api,
	}
}

// NewMutex creates a mutex with the given key name.
func (c *ClusterService) NewMutex(key string) (ClusterMutex, error) {
	return cluster.NewMutex(c.api, key)
}
