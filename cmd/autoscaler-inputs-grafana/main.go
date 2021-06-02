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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/version"
	"google.golang.org/grpc"
)

func showUsage() {
	fmt.Println("usage: autoscaler-inputs-grafana [flags]")
	flag.Usage()
}

func main() {
	var dest, addr string
	flag.StringVar(&dest, "dest", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	flag.StringVar(&addr, "addr", ":3001", "the TCP address for the server to listen on")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// TODO add flag validation

	// Note: 引数は無視

	switch {
	case showHelp:
		showUsage()
		return
	case showVersion:
		fmt.Println(version.FullVersion())
		return
	default:
		if err := listenAndServe(addr, dest); err != nil {
			log.Fatal(err)
		}
	}
}

func listenAndServe(addr, coreAddr string) error {
	handler := http.DefaultServeMux

	handler.HandleFunc("/up", func(w http.ResponseWriter, req *http.Request) {
		handleFunc(coreAddr, "up", w, req)
	})
	handler.HandleFunc("/down", func(w http.ResponseWriter, req *http.Request) {
		handleFunc(coreAddr, "down", w, req)
	})
	handler.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok")) // nolint
	})

	log.Printf("autoscaler-inputs-grafana: started on %s\n", addr)
	return http.ListenAndServe(addr, handler)
}

func handleFunc(coreAddr, requestType string, w http.ResponseWriter, req *http.Request) {
	scalingReq, err := parseRequest(requestType, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if scalingReq == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requester := &requester{dest: coreAddr}
	if err := requester.send(scalingReq); err != nil {
		log.Println("[ERROR]: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok")) // nolint
}

type scalingRequest struct {
	source      string
	action      string
	groupName   string
	requestType string
}

const (
	// StateOk       State = "ok"
	// StatePaused   State = "paused"

	StateAlerting State = "alerting"

	// StatePending  State = "pending"
	// StateNoData   State = "no_data"
)

type State string

type grafanaWebhookBody struct {
	//Title       string                   `json:"title"`
	//RuleID      int                      `json:"ruleId"`
	//RuleName    string                   `json:"ruleName"`
	//RuleURL     string                   `json:"ruleUrl"`
	State State `json:"state"`
	//Message     string                   `json:"message"`
}

func parseRequest(requestType string, req *http.Request) (*scalingRequest, error) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	log.Println("webhook received:", string(dump))

	if req.Method == http.MethodPost || req.Method == http.MethodPut {
		reqData, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		var received grafanaWebhookBody
		if err := json.Unmarshal(reqData, &received); err != nil {
			return nil, err
		}
		if received.State != StateAlerting { // alerting以外は処理しない
			return nil, nil
		}

		queryStrings := req.URL.Query()
		source := queryStrings.Get("source")
		if source == "" {
			source = defaults.SourceName
		}
		action := queryStrings.Get("action")
		if action == "" {
			action = defaults.ActionName
		}
		groupName := queryStrings.Get("resource_group_name")
		if groupName == "" {
			groupName = defaults.ResourceGroupName
		}
		return &scalingRequest{
			source:      source,
			action:      action,
			groupName:   groupName,
			requestType: requestType,
		}, nil
	}
	return nil, nil
}

type requester struct {
	dest string
}

func (r *requester) send(scalingReq *scalingRequest) error {
	if scalingReq == nil {
		return nil
	}
	ctx := context.Background()
	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, r.dest, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	req := request.NewScalingServiceClient(conn)
	var f func(ctx context.Context, in *request.ScalingRequest, opts ...grpc.CallOption) (*request.ScalingResponse, error)

	switch scalingReq.requestType {
	case "up":
		f = req.Up
	case "down":
		f = req.Down
	default:
		return fmt.Errorf("invalid request type: %s", scalingReq.requestType)
	}
	res, err := f(ctx, &request.ScalingRequest{
		Source:            scalingReq.source,
		Action:            scalingReq.action,
		ResourceGroupName: scalingReq.groupName,
	})
	if err != nil {
		return err
	}
	fmt.Printf("status: %s, job-id: %s\n", res.Status, res.ScalingJobId)
	return nil
}
