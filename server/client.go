// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

// FormatCSV requests the export to be in CSV format.
const FormatCSV = "csv"

// Client is the programmatic interface to the channel export plugin.
type Client struct {
	httpClient *http.Client
	Address    string
	AuthToken  string
	AuthType   string
}

// NewClient creates a client to the channel export plugin at the given address.
func NewClient(address string) *Client {
	return &Client{
		Address:    address,
		httpClient: &http.Client{},
	}
}

// NewMattermostServerClient creates a client to the channel export plugin at the given Mattermost server address.
//
//nolint:unused
func NewMattermostServerClient(mattermostServerAddress string) *Client {
	if !strings.HasSuffix(mattermostServerAddress, "/") {
		mattermostServerAddress += "/"
	}
	mattermostServerAddress += manifest.Id + "/"

	return NewClient(mattermostServerAddress)
}

// SetToken configures the authentication token required to identify the Mattermost user.
func (c *Client) SetToken(token string) {
	c.AuthToken = token
	c.AuthType = model.HeaderBearer
}

func (c *Client) buildURL(urlPath string, args ...any) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.Address, "/"), strings.TrimLeft(fmt.Sprintf(urlPath, args...), "/"))
}

func (c *Client) doGet(u string) (*http.Response, error) {
	r, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	if c.AuthToken != "" {
		r.Header.Set(model.HeaderAuth, c.AuthType+" "+c.AuthToken)
	}

	return c.httpClient.Do(r)
}

// ExportChannel exports the given channel in the given format to the given writer.
func (c *Client) ExportChannel(w io.Writer, channelID string, format string) error {
	u, err := url.Parse(c.buildURL("/api/v1/export"))
	if err != nil {
		return err
	}

	q := u.Query()
	q.Add("channel_id", channelID)
	q.Add("format", format)
	u.RawQuery = q.Encode()

	resp, err := c.doGet(u.String())
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			return errors.Wrap(err, "failed to copy response")
		}
		return nil

	default:
		decoder := json.NewDecoder(resp.Body)

		var apiError APIError
		err := decoder.Decode(&apiError)
		if err != nil || apiError.StatusCode != resp.StatusCode {
			return errors.Errorf("failed with status code %d", resp.StatusCode)
		}

		return &apiError
	}
}
