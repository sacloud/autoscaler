// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AutoScaler Inputs: Direct
//
// Flags:
//  -dest: (optional) URL of gRPC endpoint of AutoScaler Core. default:`unix:autoscaler.sock`
//  -action: (optional) Name of the action to perform. default:`default`
//  -group: (optional) Name of the target resource group. default:`default`
//  -source: (optional) A string representing the request source, passed to AutoScaler Core. default:`default`
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/version"
)

func main() {
	var dest, action, group, source string
	flag.StringVar(&dest, "dest", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	flag.StringVar(&action, "action", defaults.ActionName, "Name of the action to perform")
	flag.StringVar(&group, "group", defaults.ResourceGroupName, "Name of the target resource group")
	flag.StringVar(&source, "source", defaults.SourceName, "A string representing the request source, passed to AutoScaler Core")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// TODO validate

	switch {
	case showHelp:
		flag.Usage()
		return
	case showVersion:
		fmt.Println(version.FullVersion())
		return
	default:
		// TODO implements here
		log.Println("not implemented yet")
		os.Exit(1)
	}
}
