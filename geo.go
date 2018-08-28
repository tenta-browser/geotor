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
 * geo.go: Geo interface functionality
 */

package geotor

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/oschwald/maxminddb-golang"
	"github.com/sirupsen/logrus"
	"github.com/tenta-browser/polychromatic"
	"io/ioutil"
	"net"
	"path/filepath"
)

type responsewrapper struct {
	response *GeoLocation
	err      error
}

type Query struct {
	dummy bool
	ip    net.IP
	resp  chan *responsewrapper
	valid bool
}

var ErrRequestTimeout = errors.New("unable to queue the geo request for processing")

type Geo struct {
	loaded   bool
	reload   chan bool
	queries  chan *Query
	citydb   *maxminddb.Reader
	ispdb    *maxminddb.Reader
	tordb    *TorHash
	newtordb chan *TorHash
	lg       *logrus.Entry
	rt       *runtime
}

func StartGeo(cfg Config) *Geo {
	g := new(Geo)

	rt := newRuntime()

	g.lg = polychromatic.GetLogger("geo")
	g.rt = rt
	g.loaded = false
	g.reload = make(chan bool, 2) // Startup reload + after the updater runs, we might have one pending
	g.queries = make(chan *Query, 1024)
	g.newtordb = make(chan *TorHash, 1)

	go torupdater(cfg, rt, g)
	go geoupdater(cfg, rt, g)
	go geolisten(cfg, rt, g)

	return g
}

func (g *Geo) Shutdown() {
	defer g.rt.wg.Wait()

	for i := 0; i < runtimeStartedServices; i += 1 {
		g.rt.stop <- true
	}
}

func (g *Geo) Loaded() bool {
	return g.loaded
}

func (g *Geo) Query(ip net.IP) (*Query, error) {
	q := new(Query)
	q.ip = ip
	q.resp = make(chan *responsewrapper, 1) // Make sure we can stuff one in and drop it if we're already running when it gets canceled
	q.valid = true

	select {
	case g.queries <- q:
		return q, nil
	default:
		return nil, ErrRequestTimeout
	}
}

func (q *Query) Response(ctx context.Context) (*GeoLocation, error) {
	if ctx.Err() != nil {
		q.valid = false
		return nil, ctx.Err()
	}
	select {
	case wrap := <-q.resp:
		return wrap.response, wrap.err
	case <-ctx.Done():
		q.valid = false
		return nil, ctx.Err()
	}
}

func geolisten(cfg Config, rt *runtime, g *Geo) {
	defer rt.wg.Done()
	rt.wg.Add(1)
	g.lg.Debug("Started listener")
	defer func() {
		g.lg.Debug("Shut down")
	}()
	for {
		if g.loaded {
			select {
			case q := <-g.queries:
				if q.valid {
					go doQuery(q, g.citydb, g.ispdb, g.tordb, g.lg)
				} else {
					g.lg.Debug("Query is no longer valid")
				}
			case <-g.reload:
				g.lg.Debug("Got a command to reload in loaded state")
				doReload(cfg, g)
			case th := <-g.newtordb:
				g.lg.Debug("Got a new tordb in unloaded state")
				g.tordb = th
			case <-rt.stop:
				g.lg.Debug("Got shutdown command in loaded state")
				return
			}
		} else {
			select {
			case <-g.reload:
				g.lg.Debug("Got a command to reload in unloaded state")
				doReload(cfg, g)
			case th := <-g.newtordb:
				g.lg.Debug("Got a new tordb in unloaded state")
				g.tordb = th
			case <-rt.stop:
				g.lg.Debug("Got shutdown command in unloaded state")
				return
			}
		}
	}
}

func doReload(cfg Config, g *Geo) {
	g.loaded = false
	g.lg.Info("Doing a reload")
	success := 0

	var cityver, ispver string
	versionfile := filepath.Join(cfg.GeoDBPath, versionDataFilename)
	verbytes, err := ioutil.ReadFile(versionfile)
	verinfo := &VersionData{}
	if err == nil {
		err = gob.NewDecoder(bytes.NewReader(verbytes)).Decode(verinfo)
		if err == nil {
			cityver = verinfo.City
			ispver = verinfo.Isp
		}
	}

	if cityver != "" {
		cityfile := filepath.Join(cfg.GeoDBPath, fmt.Sprintf("%s-%s.mmdb", "GeoIP2-City", cityver))
		g.lg.Debugf("Geo: Opening city file %s", cityfile)
		r, err := maxminddb.Open(cityfile)
		if err == nil {
			g.citydb = r
			success += 1
		} else {
			g.lg.Errorf("Failed to open city database %s: %s", cityfile, err.Error())
		}
	}

	if ispver != "" {
		ispfile := filepath.Join(cfg.GeoDBPath, fmt.Sprintf("%s-%s.mmdb", "GeoIP2-ISP", ispver))
		g.lg.Debugf("Opening isp file %s", ispfile)
		r, err := maxminddb.Open(ispfile)
		if err == nil {
			g.ispdb = r
			success += 1
		} else {
			g.lg.Errorf("Failed to open isp database %s: %s", ispfile, err.Error())
		}
	}

	if success == 2 {
		g.loaded = true
		g.lg.Info("Reloaded Successfully")
	} else {
		g.lg.Error("Reload failure")
	}
}

func shouldIncludeSubdivision(iso string) bool {
	if iso == "US" || iso == "CA" || iso == "MX" || iso == "IN" || iso == "CN" {
		return true
	}
	return false
}

func doQuery(q *Query, citydb, ispdb *maxminddb.Reader, tordb *TorHash, lg *logrus.Entry) {
	ret := &GeoLocation{
		ISP:          &ISP{},
		LocationI18n: make(map[string]string, 0),
	}

	ispErr := ispdb.Lookup(q.ip, &ret.ISP)
	if ispErr != nil {
		lg.Warnf("ISP error: %s", ispErr.Error())
		ret.ISP = nil
	}
	var record struct {
		Position Position `maxminddb:"location"`
		City     struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"city"`
		Subdivisions []struct {
			Names map[string]string `maxminddb:"names"`
		} `maxminddb:"subdivisions"`
		Country struct {
			Names   map[string]string `maxminddb:"names"`
			ISOCode string            `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}
	lookupError := citydb.Lookup(q.ip, &record)
	if lookupError != nil {
		lg.Warnf("Lookup error: %s", lookupError.Error())
		ret = nil
	} else {

		ret.Position = &record.Position

		if country, ok := record.Country.Names["en"]; ok {
			ret.Country = country
			for lang, countryName := range record.Country.Names {
				var subdivision = ""
				if shouldIncludeSubdivision(record.Country.ISOCode) && len(record.Subdivisions) > 0 {
					subdivision, ok = record.Subdivisions[0].Names[lang]
					if !ok {
						subdivision, ok = record.Subdivisions[0].Names["en"]
						if !ok {
							subdivision = ""
						}
					}
				}
				var cityName = ""
				cityName, ok = record.City.Names[lang]
				if !ok {
					cityName, ok = record.City.Names["en"]
					if !ok {
						cityName = ""
					}
				}
				if subdivision != "" {
					countryName = fmt.Sprintf("%s, %s", subdivision, countryName)
				}
				if cityName != "" {
					countryName = fmt.Sprintf("%s, %s", cityName, countryName)
				}
				ret.LocationI18n[lang] = countryName
			}
		}
		if city, ok := record.City.Names["en"]; ok {
			ret.City = city
		}
		if location, ok := ret.LocationI18n["en"]; ok {
			ret.Location = location
		}
		ret.CountryISO = record.Country.ISOCode
	}

	if tordb != nil {
		if nodeid, present := tordb.Exists(q.ip); present {
			ret.TorNode = &nodeid
		} else {
			ret.TorNode = nil
		}
	}

	wrap := &responsewrapper{
		response: ret,
		err:      lookupError,
	}
	q.resp <- wrap
}
