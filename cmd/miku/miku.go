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

// Package main implements the CLI wrapper for the miku discord bot.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/FedorLap2006/disgolf"
	"github.com/bwmarrin/discordgo"
	"github.com/charmbracelet/log"
	"github.com/jaredallard/miku/internal/handler"
	"golang.org/x/term"
)

// Version is current version of miku.
var Version = "dev"

// main implements the miku CLI.
func main() {
	token := os.Getenv("MIKU_DISCORD_TOKEN")
	channelID := os.Getenv("MIKU_DISCORD_CHANNEL_ID")
	logFormat := os.Getenv("MIKU_LOG_FORMAT")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt, syscall.SIGSEGV)
	defer cancel()

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	// If we're a terminal AND log format wasn't set to text, default to
	// JSON. Otherwise, if log format is set to JSON, use JSON.
	if (!term.IsTerminal(int(os.Stderr.Fd())) && logFormat != "text") || logFormat == "json" {
		logger.SetFormatter(log.JSONFormatter)
	}

	logger.With("app.version", Version).Info("starting miku")

	bot, err := disgolf.New(token)
	if err != nil {
		log.With("err", err).Fatal("failed to create bot")
	}
	bot.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	h := handler.New(&handler.Config{
		ChannelID: channelID,
	}, logger)

	// Setup the main handler.
	bot.AddHandler(h.EventHandler)

	logger.Info("starting bot")
	if err := bot.Open(); err != nil {
		log.With("err", err).Fatal("failed to start bot")
	}
	defer bot.Close()

	if err := bot.UpdateWatchStatus(0, "for music links"); err != nil {
		logger.With("err", err).Warn("failed to update listening status")
	}

	logger.Info("bot started")

	// exit on signals
	<-ctx.Done()

	logger.Info("shutting down")
}
