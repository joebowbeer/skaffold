# Copyright 2019 The Skaffold Authors All rights reserved.
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

# This base image has to be updated manually after running `make build_deps`
FROM gcr.io/k8s-skaffold/build_deps:4952c2746b637e4243e5d49dc88f729ff1cf22c6 as builder
WORKDIR /skaffold
COPY . .

FROM builder as release
ARG VERSION
RUN make clean out/skaffold VERSION=$VERSION && mv out/skaffold /usr/bin/skaffold
RUN rm -rf secrets $SECRET cmd/skaffold/app/cmd/statik/statik.go
RUN skaffold credits -d /THIRD_PARTY_NOTICES
