#
# Copyright 2017-2021 The Usacloud Authors
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
GO_FILES        ?=$(shell find . -name '*.go')
AUTHOR          ?="The sacloud Authors"
COPYRIGHT_YEAR  ?="2021"
COPYRIGHT_FILES ?=$$(find . \( -name "*.dockerfile" -or -name "*.go" -or -name "*.sh" -or -name "*.pl" -or -name "*.bats" -or -name "*.bash" \) -print | grep -v "/vendor/")
BUILD_LDFLAGS   ?= "-s -w -X github.com/sacloud/autoscaler/version.Revision=`git rev-parse --short HEAD`"

export GOPROXY=https://proxy.golang.org

.PHONY: default
default: set-license fmt goimports lint test build

.PHONY: run
run:
	go run $(CURDIR)/main.go $(ARGS)

.PHONY: clean
clean:
	rm -Rf bin/*

.PHONY: tools
tools:
	(cd tools; go install golang.org/x/tools/cmd/goimports)
	(cd tools; go install golang.org/x/tools/cmd/stringer)
	(cd tools; go install github.com/sacloud/addlicense)
	(cd tools; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc)
	(cd tools; go install google.golang.org/protobuf/cmd/protoc-gen-go)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.40.0/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.40.0

.PHONY: gen
gen: gen-request gen-handler
gen-request:
	(cd docs; protoc --go_out=../request --go_opt=paths=source_relative --go-grpc_out=../request --go-grpc_opt=paths=source_relative request.proto)

gen-handler:
	(cd docs; protoc --go_out=../handler --go_opt=paths=source_relative --go-grpc_out=../handler --go-grpc_opt=paths=source_relative handler.proto)

.PHONY: build build-all build-autoscaler build-inputs-direct

build: build-autoscaler build-inputs-direct build-handlers-fake

build-autoscaler: bin/autoscaler
bin/autoscaler: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler cmd/autoscaler/main.go

build-inputs-direct: bin/autoscaler-inputs-direct
bin/autoscaler-inputs-direct: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler-inputs-direct cmd/autoscaler-inputs-direct/main.go

build-handlers-fake: bin/autoscaler-handlers-fake
bin/autoscaler-handlers-fake: $(GO_FILES)
	GOOS=$${OS:-"`go env GOOS`"} GOARCH=$${ARCH:-"`go env GOARCH`"} CGO_ENABLED=0 go build -ldflags=$(BUILD_LDFLAGS) -o bin/autoscaler-handlers-fake cmd/autoscaler-handlers-fake/main.go

.PHONY: shasum
shasum:
	(cd bin/; shasum -a 256 * > autoscaler_SHA256SUMS)

.PHONY: test
test: 
	go test $(TESTARGS) -v ./...

.PHONY: lint
lint:
	golangci-lint run ./... --modules-download-mode=readonly

.PHONY: goimports
goimports:
	goimports -l -w .

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)

.SUFFIXES:
.SUFFIXES: .go