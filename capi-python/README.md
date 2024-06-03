# Python binding for C APIs


# Useful Commands

docker run -it -v ${PWD}:/ovms --rm --entrypoint bash -p 9178:9178 openvino/model_server-build:latest

bazel build //src:ovms
find / -type f -name "pyovms.so"
find / -type f -name "ovmspybind.so"


rm /ovms/src/python/capi-python/ovmspybind.so && cp /root/.cache/bazel/_bazel_root/bc57d4817a53cab8c785464da57d1983/execroot/ovms/bazel-out/k8-opt/bin/src/python/capi-python/ovmspybind.so src/python/capi-python/

python3 src/python/capi-python/ovms-pybind.py 

bazel --output_base=$PWD/output build //src/python/capi-python:ovmspybind.so