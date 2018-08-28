/**
 * GeoTor
 *
 *    Copyright 2018 Tenta, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * For any questions, please contact developer@tenta.io
 *
 * config.go: Configuration structures
 */

package geotor

import "time"

const versionDataFilename = "geotor.version"

type Config struct {
	GeoDBPath             string
	MaxMindUrlTemplate    string
	MaxMindKey            string
	TorUrl                string
	MaxMindUpdateInterval time.Duration
	TorUpdateInterval     time.Duration
}

// NewDefaultConfig creates a sane config with daily maxmind checks and hourly tor checks
func NewDefaultConfig() Config {
	return Config{
		GeoDBPath:             "/tmp",
		MaxMindUrlTemplate:    "https://download.maxmind.com/app/geoip_download?edition_id=%s&suffix=%s&license_key=%s",
		MaxMindKey:            "",
		TorUrl:                "https://check.torproject.org/exit-addresses",
		MaxMindUpdateInterval: time.Hour * 24,
		TorUpdateInterval:     time.Hour,
	}
}

type VersionData struct {
	City string
	Isp  string
}
