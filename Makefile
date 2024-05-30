# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build-profile-launcher build-capi-python

build-profile-launcher:
	@mkdir -p ./results || true
	@cd ./profile-launcher && $(MAKE) build

build-capi-python:
	docker build -f capi_python/Dockerfile -t capi_python_bind:dev .

run:
	docker run -it --rm --entrypoint /bin/bash capi_python_bind:dev