#
# Copyright (C) 2023 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

FROM golang:1.20.9-alpine3.18 AS builder
RUN apk update && \
    apk add make autoconf libtool protobuf-dev && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY . .
RUN go mod tidy && make build-binary && make test

FROM scratch as bin
COPY --from=builder /app/profile-launcher /

FROM scratch as unittest
COPY --from=builder /app/test.out /