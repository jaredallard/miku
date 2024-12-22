// Copyright (C) 2024 miku contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0

package tidal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	"golang.org/x/oauth2/clientcredentials"
)

// tidalAPI is the base URL for the TIDAL API.
var tidalAPI = &url.URL{
	Scheme: "https",
	Host:   "openapi.tidal.com",
}

// client is a TIDAL API client.
type client struct {
	client *http.Client

	log *log.Logger
}

// newClient creates a new client using the provided clientID and
// clientSecret.'
//
//nolint:gocritic,unparam // Why: OK w/ shadow, match interface.
func newClient(ctx context.Context, clientID, clientSecret string, log *log.Logger) (*client, error) {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://auth.tidal.com/v1/oauth2/token",
	}

	return &client{config.Client(ctx), log}, nil
}

func (c *client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "github.com/jaredallard/miku")
	req.Header.Set("Accept", "application/vnd.tidal.v1+json")
	req.Header.Set("Content-Type", "application/vnd.tidal.v1+json")

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	logWith := c.log.With(
		"response.status", resp.StatusCode,
		"request.method", req.Method,
		"request.url", req.URL.String(),
	)
	defer func() {
		logWith.Debug("request to TIDAL API completed")
	}()

	// Accept any 2xx status code.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read up to 5KB of the response body to get more details.
		b, err := io.ReadAll(io.LimitReader(resp.Body, 1024*5))
		if err == nil {
			logWith = logWith.With("response.body", string(b))
		}

		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *client) get(ctx context.Context, urlStr string, params map[string]string) (*http.Response, error) {
	reqURL := tidalAPI.JoinPath(urlStr)

	// add the user-provided params to the query string.
	reqQuery := reqURL.Query()
	for k, v := range params {
		reqQuery.Add(k, v)
	}
	reqURL.RawQuery = reqQuery.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req)
}

// GetByISRC returns a resource by its ISRC. If multiple results are
// returned from the API, only the first result is returned.
func (c *client) GetByISRC(ctx context.Context, isrc, countryCode string) (*Resource, error) {
	resp, err := c.get(ctx, "/tracks/byIsrc", map[string]string{
		"isrc":        isrc,
		"countryCode": countryCode,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var l List
	if err := json.NewDecoder(resp.Body).Decode(&l); err != nil {
		return nil, err
	}

	if len(l.Data) == 0 {
		return nil, fmt.Errorf("no tracks found")
	}

	if len(l.Data) > 1 {
		c.log.Debug("multiple tracks found, using first one")
	}

	return l.Data[0].Resource, nil
}

// GetByID returns a resource by its ID.
func (c *client) GetByID(ctx context.Context, id, countryCode string) (*Resource, error) {
	resp, err := c.get(ctx, "/tracks/"+id, map[string]string{
		"countryCode": countryCode,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r ResourceContainer
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return r.Resource, nil
}
