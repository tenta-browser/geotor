Polychromatic
=============

[![Go Report Card](https://goreportcard.com/badge/github.com/tenta-browser/geotor)](https://goreportcard.com/report/github.com/tenta-browser/geotor)
[![GoDoc](https://godoc.org/github.com/tenta-browser/geotor?status.svg)](https://godoc.org/github.com/tenta-browser/geotor)

GeoTor provides a high performance, async, auto updating interface to MaxMind's GeoIP databases, and further augments this data
with up to date data from the Tor project about if nodes are tor members.

Contact: developer@tenta.io

Installation
------------

`go get github.com/tenta-browser/geotor`

Usage
-----

Call `StartGeo` to startup a geo gorouting as well as tor and geodb updaters. Call `Shutdown` (blocking) before exiting.
It will be necessary to specify at the very least the MaxMind API key in the config struct passed to StartGeo.

Call `Geo.Query(net.IP)` to perform an async query, which will be available from the returned `Query` object.

Performance
-----------

This is a high performance component designed for heavy duty production use. We use it in Tenta DNS. The async API is not
as easy to use as the underlying maxminddb package which it relies on, but allows programs to make progress while a goroutine
performs the lookup.

License
-------

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

For any questions, please contact developer@tenta.io

Contributing
------------

We welcome contributions, feedback and plain old complaining. Feel free to open
an issue or shoot us a message to developer@tenta.io. If you'd like to contribute,
please open a pull request and send us an email to sign a contributor agreement.

About Tenta
-----------

This geo library is brought to you by Team Tenta. Tenta is your [private, encrypted browser](https://tenta.com) that protects your data instead of selling. We're building a next-generation browser that combines all the privacy tools you need, including built-in OpenVPN. Everything is encrypted by default. That means your bookmarks, saved tabs, web history, web traffic, downloaded files, IP address and DNS. A truly incognito browser that's fast and easy.
