# Copyright © 2020-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#      http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG BASEIMAGE
ARG GOIMAGE
ARG VERSION="1.14.0"

# Build the sdk binary
FROM $GOIMAGE as builder

# Set envirment variable
ENV APP_NAME karavi-metrics-powerflex
ENV CMD_PATH cmd/metrics-powerflex/main.go

# Copy application data into image
COPY . /go/src/$APP_NAME
WORKDIR /go/src/$APP_NAME

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /go/src/service /go/src/$APP_NAME/$CMD_PATH

# Build the sdk image
FROM $BASEIMAGE as final
ARG VERSION
LABEL vendor="Dell Technologies" \
      maintainer="Dell Technologies" \
      name="csm-metrics-powerflex" \
      summary="Dell Container Storage Modules (CSM) for Observability - Metrics for PowerFlex" \
      description="Provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell PowerFlex" \
      release="1.16.0" \
      version=$VERSION \
      license="Apache-2.0"
COPY /licenses /licenses
COPY --from=builder /go/src/service /
ENTRYPOINT ["/service"]
