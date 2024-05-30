FROM openvino/model_server:2024.1-gpu

USER root

RUN apt-get update -y || true; DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential \
    git \
    apt-transport-https \
    curl \
    gnupg \
    python3 \
    python3-pip \
    libboost-python-dev \
    python3-pybind11 \
    python3-opencv && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set up Bazel
ENV BAZEL_VERSION 6.0.0
RUN curl -fsSL https://bazel.build/bazel-release.pub.gpg | gpg --dearmor >bazel-archive-keyring.gpg && \
    mv bazel-archive-keyring.gpg /usr/share/keyrings && \ 
    echo "deb [arch=amd64 signed-by=/usr/share/keyrings/bazel-archive-keyring.gpg] https://storage.googleapis.com/bazel-apt stable jdk1.8" | tee /etc/apt/sources.list.d/bazel.list && \
    apt update && \
    apt install bazel-6.0.0 -y

WORKDIR /
RUN git clone https://github.com/openvinotoolkit/model_server.git

WORKDIR /model_server

COPY capi_python/ /model_server/capi_python
ARG debug_bazel_flags="--strip=always --define MEDIAPIPE_DISABLE=0 --define PYTHON_DISABLE=0 --//:distro=ubuntu"
ARG JOBS=8
RUN bazel-6.0.0 build --jobs=$JOBS ${debug_bazel_flags} //:ovms_dependencies
# RUN bazel-6.0.0 --output_base=$PWD/output build //src/capi_python:ovmspybind.so