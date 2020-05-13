package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-channel-export/server/pluginapi"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/sirupsen/logrus"
)

// Handler encapsulates the context necessary for the channel export API.
type Handler struct {
	client            *pluginapi.Wrapper
	makePostsIterator func(*model.Channel) PostIterator
}

// registerAPI registers the API against the given router.
func registerAPI(router *mux.Router, client *pluginapi.Wrapper, makePostsIterator func(channel *model.Channel) PostIterator) {
	handler := &Handler{
		client:            client,
		makePostsIterator: makePostsIterator,
	}

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(mattermostAuthorizationRequired)
	api.HandleFunc("/export", handler.Export)
}

// APIError is a type of error returned by the API.
type APIError struct {
	StatusCode int
	StatusText string
	Message    string
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

func (h *Handler) hasPermissionToChannel(userID, channelID string) (*model.Channel, bool) {
	channel, err := h.client.Channel.Get(channelID)
	if appErr, ok := err.(*model.AppError); ok && appErr.StatusCode == http.StatusNotFound {
		return nil, false
	} else if err != nil {
		logrus.Warnf("failed to query channel '%s'", channelID)
		return nil, false
	}

	if h.client.User.HasPermissionToChannel(userID, channelID, model.PERMISSION_READ_CHANNEL) {
		return channel, true
	}

	return nil, false
}

// Export handles /api/v1/export, exporting the requested channel.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
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
	channel, ok := h.hasPermissionToChannel(userID, channelID)
	if !ok {
		handleError(w, http.StatusNotFound, "channel '%s' not found or user does not have permission", channelID)
		return
	}

	postIterator := h.makePostsIterator(channel)

	exporter := CSV{}
	fileName := exporter.FileName(channel.Name)

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", exporter.ContentType())
	exporter.Export(postIterator, w)
}
