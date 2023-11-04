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

// Package applemusic implements a streamingproviders.Provider for the
// Apple Music streaming service.
package applemusic

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jaredallard/miku/internal/streamingproviders"
	goapplemusic "github.com/minchao/go-apple-music"
)

var _ streamingproviders.Provider = &Provider{}

type Provider struct {
	client *goapplemusic.Client
}

// New returns a new streamingprovider.Provider for Apple Music.
// Requires the following environment variables:
// - MIKU_APPLE_MUSIC_API_TOKEN
func New(ctx context.Context) (streamingproviders.Provider, error) {
	token := os.Getenv("MIKU_APPLE_MUSIC_API_TOKEN")

	tp := goapplemusic.Transport{Token: token}
	client := goapplemusic.NewClient(tp.Client())
	return &Provider{client}, nil
}

// String returns the name of the provider.
func (p *Provider) String() string {
	return "applemusic"
}

func (p *Provider) Emoji() discordgo.ComponentEmoji {
	return discordgo.ComponentEmoji{
		Name: "miku_" + p.String(),
		ID:   "1170380264711667822",
	}
}

// LookupSongByURL returns a song from the provided URL. URL format
// should be:
// https://music.apple.com/us/album/album-name/123456789?i=123456789
func (p *Provider) LookupSongByURL(ctx context.Context, urlStr string) (*streamingproviders.Song, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	if u.Host != "music.apple.com" {
		return nil, fmt.Errorf("invalid host: %s", u.Host)
	}

	id := u.Query().Get("i")
	if id == "" {
		return nil, fmt.Errorf("missing 'i' query parameter")
	}

	// Storefront is the first part of the URL.
	storefront := strings.Split(u.Path, "/")[1]

	songs, _, err := p.client.Catalog.GetSong(ctx, storefront, id, &goapplemusic.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}
	if len(songs.Data) == 0 {
		return nil, fmt.Errorf("no songs returned")
	}
	if len(songs.Data) > 1 {
		return nil, fmt.Errorf("more than one song returned, not sure how to handle this (yet)")
	}

	// Use the first song.
	song := songs.Data[0]

	// Crude attempt at getting a 100x100 image. Not sure why they force
	// you to set the size...
	artworkURL := strings.Replace(song.Attributes.Artwork.URL, "{w}", "100", 1)
	artworkURL = strings.Replace(artworkURL, "{h}", "100", 1)

	return &streamingproviders.Song{
		ProviderEmoji: p.Emoji(),
		Provider:      p.String(),
		ProviderURL:   urlStr,
		ISRC:          song.Attributes.ISRC,
		Title:         song.Attributes.Name,
		Artists:       []string{song.Attributes.ArtistName},
		Album:         song.Attributes.AlbumName,
		AlbumArtURL:   artworkURL,
	}, nil
}

// Search returns a song from this provider using a Song provided
// from another provider.
func (p *Provider) Search(ctx context.Context, song *streamingproviders.Song) (*streamingproviders.Song, error) {
	return nil, fmt.Errorf("not implemented")
}
