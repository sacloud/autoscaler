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

// AutoScaler Core
//
// Usage:
//   autoscaler [flags]
//
// Flags:
//   -address: (optional) URL of gRPC endpoint of AutoScaler Core. default:`unix:autoscaler.sock`
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/sacloud/autoscaler/core"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/version"
)

func main() {
	var address, configPath string
	flag.StringVar(&address, "address", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	flag.StringVar(&configPath, "config", defaults.CoreConfigPath, "File path of configuration of AutoScaler Core")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// TODO add flag validation

	switch {
	case showHelp:
		showUsage()
		return
	case showVersion:
		fmt.Println(version.FullVersion())
		return
	default:
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		if err := core.Start(ctx, configPath); err != nil {
			log.Fatal(err)
		}
	}
}

func showUsage() {
	fmt.Println("usage: autoscaler [flags]")
	flag.Usage()
}
