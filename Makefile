.PHONY: all
all: help

help:
	@echo
	@echo "The following targets are commonly used:"
	@echo
	@echo "build    - Builds the code locally"
	@echo "clean    - Cleans the local build"
	@echo "podman   - Builds Podman images"
	@echo "push     - Pushes Podman images to a registry"
	@echo "check    - Runs code checking tools: lint, format, gosec, and vet"
	@echo "test     - Runs the unit tests"
	@echo

.PHONY: build
build: generate
	@$(foreach svc,$(shell ls cmd), CGO_ENABLED=0 GOOS=linux go build -o ./cmd/${svc}/bin/service ./cmd/${svc}/;)

.PHONY: clean
clean:
	rm -rf cmd/*/bin

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test -count=1 -cover -race -timeout 30s -short ./...

.PHONY: download-csm-common
download-csm-common:
	curl -O -L https://raw.githubusercontent.com/dell/csm/main/config/csm-common.mk

.PHONY: podman
podman: download-csm-common
	$(eval include csm-common.mk)
	podman build --pull $(NOCACHE) -t csm-metrics-powerflex -f Dockerfile --build-arg BASEIMAGE=$(CSM_BASEIMAGE) --build-arg GOIMAGE=$(DEFAULT_GOIMAGE) .

.PHONY: podman-no-cache
podman-no-cache:
	@make podman NOCACHE=--no-cache

.PHONY: push
push:
	podman push ${DOCKER_REPO}/csm-metrics-powerflex\:latest

.PHONY: tag
tag:
	podman tag csm-metrics-powerflex\:latest ${DOCKER_REPO}/csm-metrics-powerflex\:latest

.PHONY: check
check:
	./scripts/check.sh ./cmd/... ./opentelemetry/... ./internal/...
