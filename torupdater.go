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
 * torupdater.go: Tor node list updater
 */

package geotor

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/tenta-browser/polychromatic"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	stateNode = iota
	statePublished
	stateUpdated
	stateAddress
)

func torupdater(cfg Config, rt *runtime, g *Geo) {
	defer rt.wg.Done()
	rt.wg.Add(1)

	lg := polychromatic.GetLogger("torupdater")

	ticker := time.NewTicker(cfg.TorUpdateInterval)

	lg.Info("Starting up")

	for {
		lg.Info("Checking for updates")

		resp, err := http.Get(cfg.TorUrl)
		if err != nil {
			lg.Errorf("Unable to get tor list: %s", err.Error())
		} else {
			// Happy days, we got data

			nodes, err := tokenizeresponse(resp.Body)
			if err != nil {
				lg.Warnf("Got an error attempting to tokenize updates: %s", err)
			}
			resp.Body.Close()

			lg.Debugf("Successfully got %d tor nodes", len(nodes))
			hash := NewTorHash()
			for _, node := range nodes {
				hash.Add(node)
			}

			lg.Debugf("Successfully built a TorHash with %d entries", hash.Len())

			select {
			case g.newtordb <- hash:
				break
			default:
				lg.Debug("Unable to write new tor hash to geo")
			}
		}

		select {
		case <-ticker.C:
			// Nothing to do here, just loop to the top
		case <-rt.stop:
			ticker.Stop()
			lg.Info("Shutting down")
			return
		}
	}
}

/**
 * Parse a series of entries like this:
 *
 *    ExitNode 47E25A3042414FAA1D934D546FBF9E60E80678E2
 *    Published 2017-10-25 08:25:17
 *    LastStatus 2017-10-25 09:03:28
 *    ExitAddress 80.82.67.166 2017-10-25 09:08:02
 *
 * This is a straight forward state based parser with the small
 * wrinkle that a single node may have 1 _or more_ ExitAddresses,
 * so we have to scan until we see a new ExitNode or we run out
 * of data.
 */
func tokenizeresponse(body io.ReadCloser) ([]*TorNode, error) {
	scanner := bufio.NewScanner(body)

	state := stateNode
	node := NewTorNode()
	ret := make([]*TorNode, 0)

	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case statePublished:
			if strings.HasPrefix(line, "Published") {
				node.Published = parsetortime(line[10:])
				state = stateUpdated
			} else {
				return nil, errors.New("the Published prefix not detected")
			}
			break
		case stateUpdated:
			if strings.HasPrefix(line, "LastStatus") {
				node.Updated = parsetortime(line[11:])
				state = stateAddress
			} else {
				return nil, errors.New("the LastStatus prefix not detected")
			}
			break
		case stateAddress:
			if strings.HasPrefix(line, "ExitAddress") {
				parts := strings.SplitAfter(line, " ")
				ip := &ExitAddress{IP: net.ParseIP(strings.Trim(parts[1], " ")), Date: parsetortime(fmt.Sprintf("%s%s", parts[2], parts[3]))}
				node.Addresses = append(node.Addresses, *ip)
				break
			} else if strings.HasPrefix(line, "ExitNode") {
				ret = append(ret, node)
				node = NewTorNode()
				state = stateNode
				// Fallthrough
			} else {
				return nil, errors.New("no transition from address state found")
			}
			fallthrough
		case stateNode:
			if strings.HasPrefix(line, "ExitNode") {
				node.NodeId = line[9:]
				state = statePublished
			} else {
				return nil, errors.New("the ExitNode prefix not detected")
			}
			break
		default:
			return nil, errors.New("state error")
		}
	}
	// Handle the case where we got to the end of the file and we have a pending
	// node which we haven't put onto the output array yet
	if state == stateAddress {
		ret = append(ret, node)
	} else {
		// We didn't get a complete node at the end, which is still an error
		return nil, errors.New("parser ended in an incorrect state")
	}
	return ret, nil
}

func parsetortime(t string) time.Time {
	// 2017-05-06 10:02:47
	timestamp, err := time.Parse("2006-01-02 15:04:05", t)
	if err != nil {
		return time.Time{}
	}
	return timestamp
}
