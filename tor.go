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
 * tor.go: Tor node checker
 */

package geotor

import (
	"fmt"
	"net"
	"time"
)

// Type TorNode represents a single node in the tor network.
type TorNode struct {
	NodeId    string
	Published time.Time
	Updated   time.Time
	Addresses []ExitAddress
}

// Type ExitAddress represents an IP address and active time
type ExitAddress struct {
	IP   net.IP
	Date time.Time
}

func NewTorNode() *TorNode {
	return &TorNode{Addresses: make([]ExitAddress, 0)}
}

func (t TorNode) String() string {
	return fmt.Sprintf("TorNode %s with %d IPs", t.NodeId, len(t.Addresses))
}

var _ fmt.Stringer = TorNode{} // Verify that we're a stringer

// Type TorHash implements a hash structure for TorNodes. It is not thread safe for writes, but will
// tolerate concurrent readers.
type TorHash struct {
	hash map[string]*TorNode
	cnt  int
}

func NewTorHash() *TorHash {
	return &TorHash{hash: make(map[string]*TorNode), cnt: 0}
}

// Adds a TorNode to the hash
func (t *TorHash) Add(node *TorNode) {
	for _, addr := range node.Addresses {
		t.hash[addr.IP.String()] = node
		t.cnt += 1
	}
}

// Looks up the specified IP to see if it exists and returns the node id if it does.
func (t *TorHash) Exists(ip net.IP) (string, bool) {
	if node, ok := t.Lookup(ip); ok {
		return node.NodeId, true
	}
	return "", false
}

// Looks up the specified ip to see if it's a tor node. Returns a TorNode or nil and a boolean
func (t *TorHash) Lookup(ip net.IP) (*TorNode, bool) {
	if node, ok := t.hash[ip.String()]; ok {
		return node, true
	}
	return nil, false
}

// Indicates the number of entries in this hash
func (t *TorHash) Len() int {
	return t.cnt
}

func (t TorHash) String() string {
	return fmt.Sprintf("TorHash with %d entries", t.Len())
}

var _ fmt.Stringer = TorHash{} // Verify that we're a stringer
