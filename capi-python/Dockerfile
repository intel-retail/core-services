#
# Copyright (C) 2024 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

FROM openvino/model_server-build:latest AS builder

USER root

WORKDIR /
RUN git clone https://github.com/openvinotoolkit/model_server

WORKDIR /model_server

COPY capi-python/ /model_server/src/capi-python

RUN bazel build --jobs=8 //src/capi-python:ovmspybind.so

RUN mkdir -p /output && find / -name "ovmspybind.so" -exec cp "{}" /output/ovmspybind.so  \;

RUN find / -name "ovmspybind.so" -exec cp "{}" /model_server/src/capi-python/ovmspybind.so  \;
WORKDIR /model_server/src/capi-python

FROM scratch as bin
COPY --from=builder /output/ovmspybind.so /