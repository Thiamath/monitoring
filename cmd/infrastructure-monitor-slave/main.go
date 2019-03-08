/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Infrastructure monitor component for application clusters.
// Collects events from local cluster (i.e., Kubernetes), stores in
// a backend (i.e., Prometheus) and executes queries against the
// data.

package main

import (
	"github.com/nalej/infrastructure-monitor/cmd/infrastructure-monitor-slave/commands"
	"github.com/nalej/infrastructure-monitor/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
