# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: build-profile-launcher

build-profile-launcher:
	@mkdir -p ./results || true
	@cd ./profile-launcher && $(MAKE) build