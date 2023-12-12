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

package tidal

// List is returned through a search operation. It contains a list of
// ResourceContainers for the matching resources.
type List struct {
	Data     []ResourceContainer `json:"data,omitempty"`
	Metadata Metadata            `json:"metadata,omitempty"`
}

type Metadata struct {
	Requested int `json:"requested,omitempty"`
	Success   int `json:"success,omitempty"`
	Failure   int `json:"failure,omitempty"`
}

// ResourceContainer is a container for a single resource.
type ResourceContainer struct {
	ID       string    `json:"id"`
	Status   int       `json:"status"`
	Message  string    `json:"message"`
	Resource *Resource `json:"resource"`
}

// Resource represents a track, album, or etc returned from Tidal.
type Resource struct {
	ArtifactType string `json:"artifactType,omitempty"`
	ID           string `json:"id,omitempty"`
	Title        string `json:"title,omitempty"`
	Artists      []struct {
		ID      string `json:"id,omitempty"`
		Name    string `json:"name,omitempty"`
		Picture []struct {
			URL    string `json:"url,omitempty"`
			Width  int    `json:"width,omitempty"`
			Height int    `json:"height,omitempty"`
		} `json:"picture,omitempty"`
		Main bool `json:"main,omitempty"`
	} `json:"artists,omitempty"`
	Album struct {
		ID         string `json:"id,omitempty"`
		Title      string `json:"title,omitempty"`
		ImageCover []struct {
			URL    string `json:"url,omitempty"`
			Width  int    `json:"width,omitempty"`
			Height int    `json:"height,omitempty"`
		} `json:"imageCover,omitempty"`
		VideoCover []any `json:"videoCover,omitempty"`
	} `json:"album,omitempty"`
	Duration      int    `json:"duration,omitempty"`
	TrackNumber   int    `json:"trackNumber,omitempty"`
	VolumeNumber  int    `json:"volumeNumber,omitempty"`
	Isrc          string `json:"isrc,omitempty"`
	Copyright     string `json:"copyright,omitempty"`
	MediaMetadata struct {
		Tags []string `json:"tags,omitempty"`
	} `json:"mediaMetadata,omitempty"`
	Properties struct {
		Content []string `json:"content,omitempty"`
	} `json:"properties,omitempty"`
}
