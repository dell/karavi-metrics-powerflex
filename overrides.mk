# Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.
#
# Dell Technologies, Dell and other trademarks are trademarks of Dell Inc.
# or its subsidiaries. Other trademarks may be trademarks of their respective 
# owners.

IMAGE_REGISTRY?="sample_registry"
IMAGE_TAG?=$(shell date +%Y%m%d%H%M%S)
IMAGE_NAME="csm-metrics-powerflex"

# figure out if podman or docker should be used (use podman if found)
ifneq (, $(shell which podman 2>/dev/null))
export BUILDER=podman
else
export BUILDER=docker
endif

# target to print some help regarding these overrides and how to use them
overrides-help:
	@echo
	@echo "The following environment variables can be set to control the build"
	@echo
	@echo "REGISTRY    - The registry to push images to, default is: $(DEFAULT_REGISTRY)"
	@echo "              Current setting is: $(REGISTRY)"
	@echo "IMAGE_NAME  - The image name to be built, defaut is: $(DEFAULT_IMAGE_NAME)"
	@echo "              Current setting is: $(IMAGE_NAME)"
	@echo "IMAGE_TAG   - The image tag to be built, default is an empty string which will determine the tag by examining annotated tags in the repo."
	@echo "              Current setting is: $(IMAGE_TAG)"
	@echo
