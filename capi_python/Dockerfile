FROM openvino/model_server:2024.1-gpu

COPY . .

# Set up Bazel
ENV BAZEL_VERSION 6.0.0

RUN apt install apt-transport-https curl gnupg -y && \
    curl -fsSL https://bazel-release.pub.gpg | gpg --dearmor >bazel-archive-keyring.gpg && \
    mv bazel-archive-keyring.gpg /usr/share/keyrings &&\
    echo 

RUN bazel --output_base=$PWD/output build //capi_python:ovmspybind.so