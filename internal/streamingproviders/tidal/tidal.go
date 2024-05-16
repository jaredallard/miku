// Copyright (C) 2023 Jared Allard <jaredallard@users.noreply.github.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package tidal implements a streamingprovider for TIDAL.
package tidal

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/jaredallard/miku/internal/streamingproviders"
)

// _ ensures that Provider implements the streamingproviders.Provider
// interface.
var _ streamingproviders.Provider = &Provider{}

// Provider implements a streamingproviders.Provider.
type Provider struct {
	client *client
}

// New returns a new TIDAL client using the following environment
// variables:
// - MIKU_TIDAL_CLIENT_ID
// - MIKU_TIDAL_CLIENT_SECRET
func New(ctx context.Context, log *log.Logger) (streamingproviders.Provider, error) {
	clientID := os.Getenv("MIKU_TIDAL_CLIENT_ID")
	clientSecret := os.Getenv("MIKU_TIDAL_CLIENT_SECRET")

	c, err := newClient(ctx, clientID, clientSecret, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Provider{c}, nil
}

// Info returns information about this provider.
func (p *Provider) Info() streamingproviders.Info {
	return streamingproviders.Info{
		Identifier: "tidal",
		Name:       "TIDAL",
		Emoji: discordgo.ComponentEmoji{
			ID: "1184003400724131920",
		},
		URLHostname: "tidal.com",
	}
}

// songFromTrack converts a Resource to a streamingproviders.Song.
func (p *Provider) songFromTrack(r *Resource) *streamingproviders.Song {
	strArtists := make([]string, 0, len(r.Artists))
	for _, artist := range r.Artists {
		strArtists = append(strArtists, artist.Name)
	}

	var albumArtURL string
	if len(r.Album.ImageCover) > 0 {
		albumArtURL = r.Album.ImageCover[0].URL
	}

	return &streamingproviders.Song{
		Provider:    p.Info(),
		ISRC:        r.Isrc,
		ProviderURL: "https://tidal.com/browse/" + r.ArtifactType + "/" + r.ID,
		Title:       r.Title,
		Artists:     strArtists,
		Album:       r.Album.Title,
		Duration:    r.Duration,
		AlbumArtURL: albumArtURL,
	}
}

// LookupSongByURL returns a song from the provided URL. The URL must
// match the following format:
// - https://tidal.com/browse/track/115453632
// - https://tidal.com/browse/album/115453631
func (p *Provider) LookupSongByURL(ctx context.Context, u *url.URL) (*streamingproviders.Song, error) {
	trackPath, ID := path.Split(strings.TrimPrefix(u.Path, "/browse"))
	if trackPath != "/track/" {
		return nil, fmt.Errorf("invalid path: %s", trackPath)
	}

	track, err := p.client.GetByID(ctx, ID, "US")
	if err != nil {
		return nil, fmt.Errorf("failed to find track with ID %s: %w", ID, err)
	}

	return p.songFromTrack(track), nil
}

// Search returns a song from this provider using a Song provided from
// another provider.
func (p *Provider) Search(ctx context.Context, song *streamingproviders.Song) (*streamingproviders.Song, error) {
	track, err := p.client.GetByISRC(ctx, song.ISRC, "US")
	if err != nil {
		return nil, fmt.Errorf("failed to search for song: %w", err)
	}
	if track.ArtifactType != "track" {
		return nil, fmt.Errorf("invalid artifact type: %s", track.ArtifactType)
	}

	return p.songFromTrack(track), nil
}
