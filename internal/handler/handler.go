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

// Package handler contains the main Discord-related logic for handling
// messages.
package handler

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/jaredallard/miku/internal/streamingproviders"
	"github.com/jaredallard/miku/internal/streamingproviders/applemusic"
	"github.com/jaredallard/miku/internal/streamingproviders/spotify"
	"mvdan.cc/xurls/v2"
)

// TODO: move somewhere else
type Config struct {
	// ChannelID is the channel where the bot should listen to messages
	// from.
	ChannelID string
}

type Handler struct {
	c   *Config
	log *log.Logger

	sps []streamingproviders.Provider
}

// New creates a new handler with the default set of providers enabled.
func New(conf *Config, logger *log.Logger) *Handler {
	sps := make([]streamingproviders.Provider, 0)

	enabledProviders := []func(ctx context.Context) (streamingproviders.Provider, error){
		spotify.New,
		applemusic.New,
	}
	for _, provider := range enabledProviders {
		sp, err := provider(context.Background())
		if err != nil {
			logger.With("err", err).Fatal("failed to create provider")
		}

		logger.With("provider", sp).Info("enabled provider")
		sps = append(sps, sp)
	}

	return &Handler{conf, logger, sps}
}

// Handler implements a discordgo.EventHandler for handling new messages
// being sent.
func (h *Handler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	if m.ChannelID != h.c.ChannelID {
		return // Ignore things not in our testing channel.
	}

	h.log.With("message.contents", m.Content).Debug("observed message")

	urlx := xurls.Strict()
	urls := urlx.FindAllString(m.Content, -1)
	if len(urls) == 0 {
		return
	}

	h.log.With("urls", urls).Debug("found urls")

	// We only support one URL for now.
	url := urls[0]

	alts := h.findAlts(ctx, url)
	if len(alts) == 0 {
		h.log.Info("no alternatives found")
		return
	}

	for _, alt := range alts {
		h.log.With(
			"song.provider", alt.Provider,
			"song.title", alt.Title,
			"song.artists", alt.Artists,
		).Info("found alternative")
	}
}

// findAlts takes a URL and returns a list of all known songs for that
// URL across enabled providers.
func (h *Handler) findAlts(ctx context.Context, url string) []*streamingproviders.Song {
	// Determine which streaming provider the song is from first.
	var song *streamingproviders.Song
	for _, sp := range h.sps {
		h.log.Debug("checking provider", "provider", sp)

		var err error
		song, err = sp.LookupSongByURL(ctx, url)
		if err != nil {
			h.log.With("err", err).Debug("provider failed to lookup song")
		}
	}
	if song == nil {
		h.log.Infof("failed to find a streaming provider")
		return nil
	}

	h.log.With(
		"song.provider", song.Provider,
		"song.title", song.Title,
		"song.artists", song.Artists,
	).Info("found song")

	// Search all of the providers (minus the one we found it on) for the
	// song and return all of the results.
	var alts []*streamingproviders.Song
	for _, sp := range h.sps {
		if sp.String() == song.Provider {
			continue
		}

		// search for song
		h.log.With("provider", sp).Debug("searching for song")
		alt, err := sp.Search(ctx, song)
		if err != nil {
			h.log.With("err", err).Debug("failed to search for song")
			continue
		}

		alts = append(alts, alt)
	}

	return alts
}
