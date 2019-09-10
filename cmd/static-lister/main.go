/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Static lister app provides a Prometheus metrics endpoint that returns
// a static metric series with specified name and a label with the values
// read from label-file. We monitor label-file and change the values
// accordingly.

package main

import (
	"github.com/nalej/monitoring/cmd/static-lister/commands"
	"github.com/nalej/monitoring/version"
)

var MainVersion string
var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
