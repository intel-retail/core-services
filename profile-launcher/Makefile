# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build build-test unit-test build-binary

build:
	docker build --target bin --output=. .

build-test:
	cd test-profile && docker build -t test:dev .

unit-test:
	@docker build -f Dockerfile.unit-test -t go-unit-test:dev .
	@docker run --rm -v .:/test -v /var/run/docker.sock:/var/run/docker.sock go-unit-test:dev

build-binary:
	@echo "building executeable profile-launcher..."
	@go build -o profile-launcher
	@echo "done"
