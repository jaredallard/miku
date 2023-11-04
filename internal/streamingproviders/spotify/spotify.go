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

// Package spotify implements a streamingprovider for Spotify.
package spotify

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/jaredallard/miku/internal/streamingproviders"
	gospotify "github.com/zmb3/spotify/v2"
	gospotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var _ streamingproviders.Provider = &Spotify{}

type Spotify struct {
	client *gospotify.Client
}

// New returns a new Spotify client using the following environment
// variables:
// - MIKU_SPOTIFY_CLIENT_ID
// - MIKU_SPOTIFY_CLIENT_SECRET
func New(ctx context.Context) (streamingproviders.Provider, error) {
	clientID := os.Getenv("MIKU_SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("MIKU_SPOTIFY_CLIENT_SECRET")

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     gospotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	return &Spotify{gospotify.New(gospotifyauth.New().Client(ctx, token))}, nil
}

// String returns the name of the provider.
func (s *Spotify) String() string {
	return "spotify"
}

// songFromTrack converts a gospotify.FullTrack to a
// streamingproviders.Song.
func (s *Spotify) songFromTrack(t *gospotify.FullTrack) *streamingproviders.Song {
	strArtists := make([]string, 0, len(t.Artists))
	for _, artist := range t.Artists {
		strArtists = append(strArtists, artist.Name)
	}

	return &streamingproviders.Song{
		Provider:    s.String(),
		ISRC:        t.ExternalIDs["isrc"],
		ProviderURL: t.ExternalURLs["spotify"],
		Title:       t.Name,
		Artists:     strArtists,
		Album:       t.Album.Name,
	}
}

// LookupSongByURL returns a song from the provided URL. The URL must
// match the following format:
// - https://open.spotify.com/track/1qRbITa6QZoD6kQpBLMgao
func (s *Spotify) LookupSongByURL(ctx context.Context, urlStr string) (*streamingproviders.Song, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Only want spotify links.
	if u.Host != "open.spotify.com" {
		return nil, fmt.Errorf("invalid host: %s", u.Host)
	}

	trackPath, ID := path.Split(u.Path)
	if trackPath != "/track/" {
		return nil, fmt.Errorf("invalid path: %s", trackPath)
	}

	track, err := s.client.GetTrack(ctx, gospotify.ID(ID))
	if err != nil {
		return nil, fmt.Errorf("failed to find track with ID %s: %w", ID, err)
	}

	strArtists := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		strArtists = append(strArtists, artist.Name)
	}

	return s.songFromTrack(track), nil
}

// Search returns a song from this provider using a Song provided from
// another provider.
func (s *Spotify) Search(ctx context.Context, song *streamingproviders.Song) (*streamingproviders.Song, error) {
	res, err := s.client.Search(ctx, "isrc:"+song.ISRC, gospotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("failed to search for song: %w", err)
	}

	if res.Tracks == nil || res.Tracks.Tracks == nil || len(res.Tracks.Tracks) == 0 {
		return nil, fmt.Errorf("no tracks returned")
	}

	track := res.Tracks.Tracks[0]
	return s.songFromTrack(&track), nil
}
