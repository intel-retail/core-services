# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build build-binary test

build:
	docker build --target bin --output=. .

build-binary:
	@echo "building executeable profile-launcher..."
	@go build -o profile-launcher
	@echo "done"

test:
	go test -p 1 -cover ./...
