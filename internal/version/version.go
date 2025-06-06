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

// Package version contains the version of the application.
package version

import "strings"

// Version is the version of the application.
var Version = "dev"

// init populates the Version with a v prefix.
//
//nolint:gochecknoinits // Why: init is the right place for this.
func init() {
	// Ensure we always have a v prefix.
	Version = "v" + strings.TrimPrefix(Version, "v")
}
