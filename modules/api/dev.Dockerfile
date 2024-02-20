# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.22-alpine3.19

WORKDIR /workspace

RUN go install github.com/cosmtrek/air@latest

ADD . .

WORKDIR /workspace/api

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# The port that the application listens on.
EXPOSE 9000 9001
ENTRYPOINT ["air", "-c", ".air.toml", "--", "--insecure-bind-address=0.0.0.0", "--bind-address=0.0.0.0"]
