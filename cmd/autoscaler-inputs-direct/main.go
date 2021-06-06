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
// Usage:
//   autoscaler-inputs-direct [flags] up|down|status
//
// Arguments:
//   up: run the Up func
//   down: run the Down func
//
// Flags:
//   -dest: (optional) URL of gRPC endpoint of AutoScaler Core. default:`unix:autoscaler.sock`
//   -action: (optional) Name of the action to perform. default:`default`
//   -group: (optional) Name of the target resource group. default:`default`
//   -source: (optional) A string representing the request source, passed to AutoScaler Core. default:`default`
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/version"
	"google.golang.org/grpc"
)

func main() {
	var dest, action, group, source, desiredStateName string
	flag.StringVar(&dest, "dest", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	flag.StringVar(&action, "action", defaults.ActionName, "Name of the action to perform")
	flag.StringVar(&group, "group", defaults.ResourceGroupName, "Name of the target resource group")
	flag.StringVar(&source, "source", defaults.SourceName, "A string representing the request source, passed to AutoScaler Core")
	flag.StringVar(&desiredStateName, "desired-state-name", defaults.DesiredStateName, "Name of the desired state defined in Core's configuration file")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// TODO add flag validation

	if flag.NArg() != 1 {
		showUsage()
		return
	}

	command := flag.Args()[0]
	if command != "up" && command != "down" {
		showUsage()
		os.Exit(1)
	}

	switch {
	case showHelp:
		showUsage()
		return
	case showVersion:
		fmt.Println(version.FullVersion())
		return
	default:
		ctx := context.Background()
		// TODO 簡易的な実装、後ほど整理&切り出し
		conn, err := grpc.DialContext(ctx, dest, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		req := request.NewScalingServiceClient(conn)
		var f func(ctx context.Context, in *request.ScalingRequest, opts ...grpc.CallOption) (*request.ScalingResponse, error)

		switch command {
		case "up":
			f = req.Up
		case "down":
			f = req.Down
		default:
			log.Fatal("invalid args")
		}
		res, err := f(ctx, &request.ScalingRequest{
			Source:            source,
			Action:            action,
			ResourceGroupName: group,
			DesiredStateName:  desiredStateName,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("status: %s, job-id: %s", res.Status, res.ScalingJobId)
		if res.Message != "" {
			fmt.Printf(", message: %s", res.Message)
		}
		fmt.Println()
	}
}

func showUsage() {
	fmt.Println("usage: autoscaler-inputs-direct [flags] up|down")
	flag.Usage()
}
