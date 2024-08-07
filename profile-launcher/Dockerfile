#
# Copyright (C) 2023 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.21.13-alpine3.19 AS builder
RUN apk update && \
    apk add make autoconf libtool protobuf-dev && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY . .
RUN go mod tidy && make build-binary

FROM scratch as bin
COPY --from=builder /app/profile-launcher /