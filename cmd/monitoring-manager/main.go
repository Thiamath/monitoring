/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Monitoring manager component for management clusters.
// Forward all queries to requested application cluster.

package main

import (
	"github.com/nalej/monitoring/cmd/monitoring-manager/commands"
	"github.com/nalej/monitoring/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
