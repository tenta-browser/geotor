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
 * geo_test.go: Very basic test
 */

package geotor

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/tenta-browser/polychromatic"
	"math/rand"
	"net"
	"testing"
	"time"
)

func TestStartGeo(t *testing.T) {
	polychromatic.SetLogLevel(logrus.DebugLevel)
	c := NewDefaultConfig()
	c.MaxMindKey = "PUT YOURS HERE"

	g := StartGeo(c)

	for !g.Loaded() {
		time.Sleep(100 * time.Millisecond)
	}

	for i := 0; i < 100; i += 1 {
		go func() {
			b := make([]byte, 4)
			rand.Read(b)
			ip := net.IPv4(b[0], b[1], b[2], b[3])
			q, err := g.Query(ip)
			if err != nil {
				t.Fail()
			}
			r, err := q.Response(context.TODO())
			if err != nil {
				t.Fail()
			}
			tor := r.TorNode != nil
			polychromatic.GetLogger("test").Infof("%s -> %s / %v", ip.String(), r.Location, tor)
		}()
	}

	g.Shutdown()
}
