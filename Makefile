GOSCALEIO_DIR=../goscaleio
GOSCALEIO_BRANCH=ioad-with-fixes
GOSCALEIO_CLONED_BRANCH := $(shell git --git-dir=$(GOSCALEIO_DIR)/.git --work-tree=$(GOSCALEIO_DIR) rev-parse --abbrev-ref HEAD)

.PHONY: all
all: help

help:
	@echo
	@echo "The following targets are commonly used:"
	@echo
	@echo "build    - Builds the code locally"
	@echo "clean    - Cleans the local build"
	@echo "docker   - Builds Docker images"
	@echo "push     - Pushes Docker images to a registry"
	@echo "check    - Runs code checking tools: lint, format, gosec, and vet"
	@echo "test     - Runs the unit tests"
	@echo

.PHONY: build
build: generate
	@if [ ! -d $(GOSCALEIO_DIR) ];then \
		git clone --branch $(GOSCALEIO_BRANCH) https://github.com/dell/goscaleio.git $(GOSCALEIO_DIR); \
	elif [ "$(GOSCALEIO_CLONED_BRANCH)" != "$(GOSCALEIO_BRANCH)" ];then \
		git --git-dir=$(GOSCALEIO_DIR)/.git --work-tree=$(GOSCALEIO_DIR) fetch origin; \
		git --git-dir=$(GOSCALEIO_DIR)/.git --work-tree=$(GOSCALEIO_DIR) checkout $(GOSCALEIO_BRANCH); \
	fi
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

.PHONY: docker
docker:
	SERVICE=cmd/powerflex-metrics docker build -t karavi-powerflex-metrics -f Dockerfile cmd/powerflex-metrics/

.PHONY: push
push:
	docker push ${DOCKER_REPO}/karavi-powerflex-metrics\:latest

.PHONY: tag
tag:
	docker tag karavi-powerflex-metrics\:latest ${DOCKER_REPO}/karavi-powerflex-metrics\:latest

.PHONY: check
check:
	./scripts/check.sh ./cmd/... ./opentelemetry/... ./internal/...
