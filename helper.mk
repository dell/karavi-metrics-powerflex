# Copyright © 2026 Dell Inc. or its subsidiaries. All Rights Reserved.
#
# Dell Technologies, Dell and other trademarks are trademarks of Dell Inc.
# or its subsidiaries. Other trademarks may be trademarks of their respective 
# owners.

generate:
	go generate ./...

download-csm-common:
	git clone --depth 1 git@github.com:CSM/csm.git csm-temp-repo
	cp csm-temp-repo/config/csm-common.mk .
	rm -rf csm-temp-repo

vendor:
	GOPRIVATE=github.com go mod vendor
