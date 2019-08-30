/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// CLI interacting with the Availability Monitoring Platform

package main

import (
	"github.com/nalej/monitoring/cmd/availability-cli/commands"
	"github.com/nalej/monitoring/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
