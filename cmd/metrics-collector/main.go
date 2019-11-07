/*
 * Copyright 2019 Nalej
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
 */

// Monitoring component for application clusters.
// Collects events from local cluster (i.e., Kubernetes), stores in
// a backend (i.e., Prometheus) and executes queries against the
// data.

package main

import (
	"github.com/nalej/monitoring/cmd/metrics-collector/commands"
	"github.com/nalej/monitoring/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
