/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runner

import "sync"

type Intents struct {
	build      bool
	sync       bool
	deploy     bool
	autoBuild  bool
	autoSync   bool
	autoDeploy bool

	lock sync.Mutex
}

func newIntents(autoBuild, autoSync, autoDeploy bool) *Intents {
	i := &Intents{
		autoBuild:  autoBuild,
		autoSync:   autoSync,
		autoDeploy: autoDeploy,
	}

	return i
}

func (i *Intents) reset() {
	i.lock.Lock()
	i.build = i.autoBuild
	i.sync = i.autoSync
	i.deploy = i.autoDeploy
	i.lock.Unlock()
}

func (i *Intents) resetBuild() {
	i.lock.Lock()
	i.build = i.autoBuild
	i.lock.Unlock()
}

func (i *Intents) resetSync() {
	i.lock.Lock()
	i.sync = i.autoSync
	i.lock.Unlock()
}

func (i *Intents) resetDeploy() {
	i.lock.Lock()
	i.deploy = i.autoDeploy
	i.lock.Unlock()
}

func (i *Intents) setBuild(val bool) {
	i.lock.Lock()
	i.build = val
	i.lock.Unlock()
}

func (i *Intents) setSync(val bool) {
	i.lock.Lock()
	i.sync = val
	i.lock.Unlock()
}

func (i *Intents) setDeploy(val bool) {
	i.lock.Lock()
	i.deploy = val
	i.lock.Unlock()
}

func (i *Intents) getAutoBuild() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoBuild
}

func (i *Intents) getAutoSync() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoSync
}

func (i *Intents) getAutoDeploy() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoDeploy
}

func (i *Intents) setAutoBuild(val bool) {
	i.lock.Lock()
	i.autoBuild = val
	i.lock.Unlock()
}

func (i *Intents) setAutoSync(val bool) {
	i.lock.Lock()
	i.autoSync = val
	i.lock.Unlock()
}

func (i *Intents) setAutoDeploy(val bool) {
	i.lock.Lock()
	i.autoDeploy = val
	i.lock.Unlock()
}

// returns build, sync, and deploy intents (in that order)
func (i *Intents) GetIntents() (bool, bool, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.build, i.sync, i.deploy
}

func (i *Intents) IsAnyAutoEnabled() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoBuild || i.autoSync || i.autoDeploy
}
