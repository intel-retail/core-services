# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build build-test unit-test build-binary

build:
	docker build --target bin --output=. .

build-test:
	cd test-profile && docker build -t test:dev .

unit-test:
	docker build --target unittest --output=. .

build-binary:
	@echo "building executeable profile-launcher..."
	@go build -o profile-launcher
	@echo "done"

test:
	go test -cover ./... >> test.out
