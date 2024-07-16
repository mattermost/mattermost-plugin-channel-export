package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	KeyClusterMutex = "mutex_exporter"
)

// Handler encapsulates the context necessary for the channel export API.
type Handler struct {
	client            *pluginapi.Wrapper
	makePostsIterator func(*model.Channel, bool) PostIterator
	clusterMutex      pluginapi.ClusterMutex
}

// registerAPI registers the API against the given router.
func (p *Plugin) registerAPI(router *mux.Router, client *pluginapi.Wrapper, makePostsIterator func(*model.Channel, bool) PostIterator) error {
	clusterMutex, err := client.Cluster.NewMutex(KeyClusterMutex)
	if err != nil {
		return fmt.Errorf("cannot create cluster mutex: %w", err)
	}

	p.handler = Handler{
		client:            client,
		makePostsIterator: makePostsIterator,
		clusterMutex:      clusterMutex,
	}

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(mattermostAuthorizationRequired)
	api.HandleFunc("/export", p.Export)
	return nil
}

// APIError is a type of error returned by the API.
type APIError struct {
	StatusText string
	Message    string
	StatusCode int
}

func (e *APIError) Error() string {
	return e.Message
}

func handleError(w http.ResponseWriter, statusCode int, message string, a ...interface{}) {
	message = fmt.Sprintf(message, a...)
	logrus.Warnf("%s (%d): %s", http.StatusText(statusCode), statusCode, message)

	w.WriteHeader(statusCode)
	b, _ := json.Marshal(APIError{
		StatusCode: statusCode,
		StatusText: http.StatusText(statusCode),
		Message:    message,
	})
	_, err := w.Write(b)
	if err != nil {
		logrus.WithError(err).Warnf("failed to handle error")
	}
}

// mattermostAuthorizationRequired requires a Mattermost user to have authenticated.
func mattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (p *Plugin) hasPermissionToChannel(userID, channelID string) (*model.Channel, bool) {
	channel, err := p.handler.client.Channel.Get(channelID)
	if appErr, ok := err.(*model.AppError); ok && appErr.StatusCode == http.StatusNotFound {
		return nil, false
	} else if err != nil {
		logrus.Warnf("failed to query channel '%s'", channelID)
		return nil, false
	}

	if p.handler.client.User.HasPermissionToChannel(userID, channelID, model.PermissionReadChannel) {
		return channel, true
	}

	return nil, false
}

// Export handles /api/v1/export, exporting the requested channel.
func (p *Plugin) Export(w http.ResponseWriter, r *http.Request) {
	// only allow one export at a time
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	if err := p.handler.clusterMutex.LockWithContext(ctx); err != nil {
		handleError(w, http.StatusServiceUnavailable, "a channel export is already running.")
		return
	}
	defer func() {
		p.handler.clusterMutex.Unlock()
	}()

	license := p.handler.client.System.GetLicense()
	if !isLicensed(license, p.handler.client) {
		handleError(w, http.StatusBadRequest, "the channel export plugin requires a valid E20 license.")
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		handleError(w, http.StatusBadRequest, "missing channel_id parameter")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		handleError(w, http.StatusBadRequest, "missing format parameter")
		return
	}
	if format != "csv" {
		handleError(w, http.StatusBadRequest, "unsupported format parameter '%s'", format)
		return
	}

	userID := r.Header.Get("Mattermost-User-ID")
	channel, ok := p.hasPermissionToChannel(userID, channelID)
	if !ok {
		handleError(w, http.StatusNotFound, "channel '%s' not found or user does not have permission", channelID)
		return
	}

	conf := p.getConfiguration()
	if conf.EnableAdminRestrictions {
		if !(p.handler.client.User.HasPermissionTo(userID, model.PermissionManageChannelRoles) || p.handler.client.User.HasPermissionTo(userID, model.PermissionManageSystem)) {
			handleError(w, http.StatusNotFound, "user does not have permission", channelID)
			return
		}
	}

	postIterator := p.handler.makePostsIterator(channel, showEmailAddress(p.handler.client, userID))

	exporter := CSV{}
	fileName := exporter.FileName(channel.Name)

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", exporter.ContentType())
	err := exporter.Export(postIterator, w)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "failed to create the exported data")
	}
}
