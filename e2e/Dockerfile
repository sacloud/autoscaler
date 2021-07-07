# Copyright 2021 The sacloud Authors
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

FROM golang:1.16 AS builder
MAINTAINER Usacloud Authors <sacloud.users@gmail.com>

# Install dependencies
RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
        zip  \
        bzr  \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

# Install Terraform

RUN wget https://releases.hashicorp.com/terraform/1.0.0/terraform_1.0.0_linux_amd64.zip
RUN unzip terraform_1.0.0_linux_amd64.zip && rm terraform_1.0.0_linux_amd64.zip
RUN mv terraform /usr/bin/terraform

# Install Usacloud
RUN wget https://github.com/sacloud/usacloud/releases/download/v1.1.6/usacloud_linux-amd64.zip
RUN unzip usacloud_linux-amd64.zip && rm usacloud_linux-amd64.zip
RUN mv usacloud /usr/bin/usacloud

# Install grpc_health_probe

RUN go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.2

# Build sacloud/autoscaler

ADD . /go/src/github.com/sacloud/autoscaler
WORKDIR /go/src/github.com/sacloud/autoscaler

ENV CGO_ENABLED 0
RUN make tools build

RUN mv /go/src/github.com/sacloud/autoscaler/bin/autoscaler /usr/bin/autoscaler

