# Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.
#
# Dell Technologies, Dell and other trademarks are trademarks of Dell Inc.
# or its subsidiaries. Other trademarks may be trademarks of their respective 
# owners.

include images.mk

.PHONY: all
all: help

help:
	@echo
	@echo "The following targets are commonly used:"
	@echo
	@echo "build    - Builds the code locally"
	@echo "clean    - Cleans the local build"
	@echo "check    - Runs code checking tools: lint, format, gosec, and vet"
	@echo "test     - Runs the unit tests"
	@echo "vendor 	- Downloads a vendor list (local copy) of repositories required to compile the repo."
	@echo

.PHONY: build
build:
	@$(foreach svc,$(shell ls cmd), CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o ./cmd/${svc}/bin/service ./cmd/${svc}/;)

.PHONY: clean
clean:
	rm -rf cmd/*/bin
	rm -rf csm-common.mk
	rm -rf vendor

.PHONY: test
test:
	go test -count=1 -cover -race -timeout 30s -short ./...

.PHONY: check
check:
	./scripts/check.sh ./cmd/... ./opentelemetry/... ./internal/...

mockgen:
	go install go.uber.org/mock/mockgen@latest
