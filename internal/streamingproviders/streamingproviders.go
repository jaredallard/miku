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

// Package streamingproviders implements a generic interface and types
// for interacting with various Music streaming providers. Primarily
// geared towards getting song information.
package streamingproviders

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// Song is a music track.
type Song struct {
	// ProviderEmoji is the emoji to use for this this song's provider.
	// TODO: This shouldn't be here...
	ProviderEmoji discordgo.ComponentEmoji

	// Provider is the name of the provider that returned this song.
	Provider string

	// ProviderURL is the URL of the song on the provider's website. This
	// should be publicly accessible.
	ProviderURL string

	// ISRC is the international standard recording code for the song.
	// This is used to uniquely identify a song.
	ISRC string

	// Title is the title of the song.
	Title string

	// Artists is the list of artists on the song. The first artist is
	// considered the primary artist.
	Artists []string

	// Album is the album of the song.
	Album string

	// AlbumArtURL is the URL of the album art for the song. This must be
	// publicly accessible.
	AlbumArtURL string
}

// NewProvider is a function that returns a new Provider. If a provider
// is unable to be used (e.g., no authentication) it should return an
// error. Callers should handle the error and only fail if that provider
// is required, otherwise consider it disabled.
type NewProvider func(ctx context.Context) (Provider, error)

// Provider is a streaming provider interface capable of looking up
// songs by URLs or search queries and returning information about
// them.
type Provider interface {
	// String return the name of the provider.
	String() string

	// Emoji returns the emoji to use for this provider.
	Emoji() discordgo.ComponentEmoji

	// LookupSongByURL returns a song from the provided URL.
	LookupSongByURL(ctx context.Context, url string) (*Song, error)

	// Search returns a song from this provider using a Song provided from
	// another provider.
	Search(ctx context.Context, song *Song) (*Song, error)
}
