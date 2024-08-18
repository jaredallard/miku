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

// Package handler contains the main Discord-related logic for handling
// messages.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/jaredallard/miku/internal/streamingproviders"
	"github.com/jaredallard/miku/internal/streamingproviders/applemusic"
	"github.com/jaredallard/miku/internal/streamingproviders/spotify"
	"github.com/jaredallard/miku/internal/streamingproviders/tidal"
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

	enabledProviders := []func(context.Context, *log.Logger) (streamingproviders.Provider, error){
		spotify.New,
		applemusic.New,
		tidal.New,
	}
	for _, provider := range enabledProviders {
		plog := logger.With()

		sp, err := provider(context.Background(), plog)
		if err != nil {
			logger.With("err", err).Fatal("failed to create provider")
		}

		// update the logger to include the provider name
		(*plog) = *plog.With("provider.id", sp.Info().Identifier)

		plog.Info("enabled provider")
		sps = append(sps, sp)
	}

	return NewWithProviders(conf, logger, sps)
}

// NewWithProviders creates a new handler with the provided providers.
func NewWithProviders(conf *Config, logger *log.Logger, sps []streamingproviders.Provider) *Handler {
	return &Handler{conf, logger, sps}
}

// Handler implements a discordgo.EventHandler for handling new messages
// being sent.
func (h *Handler) EventHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()
	if m.ChannelID != h.c.ChannelID {
		return // Ignore things not in our channel.
	}

	if m.Author.Bot {
		return // Ignore bots.
	}

	h.log.With("message.contents", m.Content).Debug("observed message")

	urlx := xurls.Strict()
	urls := urlx.FindAllString(m.Content, -1)
	if len(urls) == 0 {
		h.log.Debug("no urls found in message")
		return
	}

	h.log.With("urls", urls).Debug("found urls")

	originalSong, alts, err := h.NewURL(ctx, urls[0])
	if err != nil {
		h.log.With("err", err).Error("failed to handle url")
		return
	}

	// Send a message back to the user.
	if err := h.sendMessage(s, m, urls, originalSong, alts); err != nil {
		h.log.With("err", err).Error("failed to send message")
		return
	}
}

// NewURL takes a URL and searches all enabled providers for it. It then
// searches all provides (minus the one the song was found on) and
// returns alternative streamingproviders where that song was found (the
// alternatives).
func (h *Handler) NewURL(ctx context.Context, url string) (*streamingproviders.Song,
	[]*streamingproviders.Song, error) {
	originalSong, alts := h.findAlts(ctx, url)
	if originalSong == nil {
		return nil, nil, fmt.Errorf("failed to find original song")
	}
	h.log.With(
		"song.isrc", originalSong.ISRC,
		"song.provider", originalSong.Provider.Identifier,
		"song.title", originalSong.Title,
		"song.artists", originalSong.Artists,
	).Info("found original song")
	if len(alts) == 0 {
		return nil, nil, fmt.Errorf("failed to find alternatives for song")
	}

	for _, alt := range alts {
		h.log.With(
			"song.provider", alt.Provider.Identifier,
			"song.title", alt.Title,
			"song.artists", alt.Artists,
		).Info("found alternative")
	}

	return originalSong, alts, nil
}

// sendMessage sends a reply to the original message with information on
// the current song as well as alternatives.
func (h *Handler) sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, urls []string,
	song *streamingproviders.Song, alts []*streamingproviders.Song) error {

	// Convert the duration into a human readable format.
	duration := fmt.Sprintf("%d:%02d", song.Duration/60, song.Duration%60)

	msg := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{{
			Type:        discordgo.EmbedTypeRich,
			Title:       song.Title,
			Description: strings.Join(song.Artists, ", "),
			URL:         song.ProviderURL,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL:    song.AlbumArtURL,
				Height: 50,
				Width:  50,
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s · Duration %s · Shared by @%s", song.Provider.Name, duration, m.Author.Username),
			},
		}},
	}

	// Remove URLs from the original message and see if there's still
	// anything left. If so, we should send it with the message.
	content := m.Content
	for _, url := range urls {
		content = strings.Replace(content, url, "", 1)
	}
	content = strings.TrimSpace(content)

	if content != "" {
		msg.Content = fmt.Sprintf(" > %s: %s", m.Author.Mention(), content)
	}

	// Create a copy of alts with the original song. We want to show it at
	// the end of the message.
	songEmbeds := append([]*streamingproviders.Song{}, alts...)
	songEmbeds = append(songEmbeds, song)

	var row []discordgo.MessageComponent
	for i := range songEmbeds {
		alt := songEmbeds[i]
		row = append(row, discordgo.Button{
			URL:   alt.ProviderURL,
			Emoji: &alt.Provider.Emoji,
			Style: discordgo.LinkButton,
		})
	}

	// We need to wrap the rows in an actionsrow component.
	msg.Components = append(msg.Components, discordgo.ActionsRow{
		Components: row,
	})

	// encode to JSON so we can debug it easier
	b, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.log.With("discord.message", string(b)).Debug("sending message")
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, msg); err != nil {
		return fmt.Errorf("failed to send reply: %w", err)
	}

	h.log.With("discord.message", string(m.Reference().MessageID)).Debug("deleting original message")
	if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
		return fmt.Errorf("failed to delete original message: %w", err)
	}

	return nil
}

// findOriginalSongByURL iterates over all enabled providers and returns
// the first song that can be found on a provider. This function will
// return nil if no song can be found.
//
// !!! IMPORTANT: Can return nil. See function definition.
func (h *Handler) findOriginalSongByURL(ctx context.Context, urlStr string) *streamingproviders.Song {
	for _, sp := range h.sps {
		pinfo := sp.Info()
		plog := h.log.With("provider.id", pinfo.Identifier)

		u, err := url.Parse(urlStr)
		if err != nil {
			plog.With("err", err).Error("failed to parse url")
			continue
		}

		// If our provider has a hostname, make sure the URL matches it.
		if pinfo.URLHostname != "" && u.Hostname() != pinfo.URLHostname {
			plog.Debug("url doesn't match provider's hostname")
			continue
		}

		plog.Debug("looking for song via URL")

		song, err := sp.LookupSongByURL(ctx, u)
		if err == nil { // Found it.
			plog.Info("found song")
			return song
		}

		plog.With("err", err).Debug("provider failed to lookup song")
	}

	// Didn't find it after searching all enabled providers.
	return nil
}

// findAlts takes a URL and returns a list of all known songs for that
// URL across enabled providers.
func (h *Handler) findAlts(ctx context.Context, url string) (*streamingproviders.Song, []*streamingproviders.Song) {
	song := h.findOriginalSongByURL(ctx, url)
	if song == nil {
		return nil, nil
	}

	// Search all of the providers (minus the one we found it on) for the
	// song and return all of the results.
	var alts []*streamingproviders.Song
	for _, sp := range h.sps {
		if sp.Info().Identifier == song.Provider.Identifier {
			continue
		}

		h.log.With("provider.id", sp.Info().Identifier).Debug("searching for alternative")
		alt, err := sp.Search(ctx, song)
		if err != nil {
			h.log.With("err", err).Debug("failed to search for song")
			continue
		}

		alts = append(alts, alt)
	}

	return song, alts
}
