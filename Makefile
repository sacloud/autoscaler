#
# Copyright 2021-2025 The sacloud/autoscaler Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#====================
AUTHOR          ?= The sacloud/autoscaler Authors
COPYRIGHT_YEAR  ?= 2021-2025

BIN            ?= bin/autoscaler
BUILD_LDFLAGS   ?= "-s -w -X github.com/sacloud/autoscaler/version.Revision=`git rev-parse --short HEAD`"

include includes/go/common.mk
include includes/go/single.mk
#====================
export GOPROXY=https://proxy.golang.org

default: $(DEFAULT_GOALS)
tools: dev-tools

.PHONY: tools
tools: dev-tools
	@echo "[INFO] please install clang-format manually if you would like to edit .proto"
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
	go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.8

.PHONY: clean-all
clean-all:
	rm -rf bin/

.PHONY: gen
gen: gen-request gen-handler
gen-request:
	(cd protos; protoc --go_out=../request --go_opt=paths=source_relative --go-grpc_out=../request --go-grpc_opt=paths=source_relative request.proto)
gen-handler:
	(cd protos; protoc --go_out=../handler --go_opt=paths=source_relative --go-grpc_out=../handler --go-grpc_opt=paths=source_relative handler.proto)

.PHONY: build-handlers-fake
build-handlers-fake: bin/autoscaler-handlers-fake
bin/autoscaler-handlers-fake: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler-handlers-fake cmd/autoscaler-handlers-fake/main.go

.PHONY: e2e-test
e2e-test: install
	@echo "[INFO] When you run e2e-test for the first time, run 'make tools' first."
	(cd e2e; go test $(TESTARGS) -v -tags=e2e -timeout 240m ./...)

.PHONY: fmt-proto
fmt-proto:
	find $(CURDIR)/protos/ -name "*.proto" | xargs clang-format -i

.SUFFIXES:
.SUFFIXES: .go
