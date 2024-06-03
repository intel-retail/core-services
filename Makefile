# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build-profile-launcher ovms-build-image build-capi-python run

build-profile-launcher:
	@mkdir -p ./results || true
	@cd ./profile-launcher && $(MAKE) build

ovms-build-image:
	git clone https://github.com/openvinotoolkit/model_server
	cd model_server && CHECK_COVERAGE=0 RUN_TESTS=0 make ovms_builder_image

build-capi-python:
	docker build -f capi-python/Dockerfile -t capi-python-bind:dev --target bin --output=. .

download-example-model:
	cd capi-python && ./downloadExampleModel.sh

run:download-example-model
	docker run -it --rm -v ./capi-python:/tmp/project --entrypoint /bin/bash capi-python-bind:dev