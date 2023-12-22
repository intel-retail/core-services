# Copyright © 2023 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: update-submodules build-profile-launcher

update-submodules:
	git submodule update --init --recursive

build-profile-launcher:
	@mkdir -p ./results || true
	@cd ./profile-launcher && $(MAKE) build