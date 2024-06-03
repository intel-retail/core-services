# FROM openvino/model_server:2024.1-gpu
FROM openvino/model_server-build:latest

USER root

WORKDIR /
RUN git clone https://github.com/openvinotoolkit/model_server

WORKDIR /model_server

COPY capi_python/ /model_server/src/capi_python

RUN bazel --jobs=8 build //src/capi_python:ovmspybind.so

RUN find / -name "ovmspybind.so" -exec cp "{}" /output/ovmspybind.so  \;

# FROM scratch as bin
# COPY --from=builder /output/ovmspybind.so /