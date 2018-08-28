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
 * runtime.go: Tooling to handle running required background goroutines
 */

package geotor

import (
	"sync"
)

const runtimeStartedServices = 3

type runtime struct {
	wg      *sync.WaitGroup
	stop    chan bool
}

func newRuntime() *runtime {
	return &runtime{
		wg:      &sync.WaitGroup{},
		stop:    make(chan bool, runtimeStartedServices),
	}
}