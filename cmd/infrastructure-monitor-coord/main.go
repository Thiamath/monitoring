/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Infrastructure monitor component for management clusters.
// Forward all queries to requested application cluster.

package main

import (
	"github.com/nalej/infrastructure-monitor/cmd/infrastructure-monitor-coord/commands"
	"github.com/nalej/golang-template/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
