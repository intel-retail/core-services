docker run -it -v ${PWD}:/ovms --rm --entrypoint bash -p 9178:9178 openvino/model_server-build:latest

bazel build //src:ovms
find / -type f -name "pyovms.so"
find / -type f -name "ovmspybind.so"
find / -name "mytest.txt" -exec cp "{}" /output/mytest.txt  \;

rm /ovms/src/python/capi_python/ovmspybind.so && cp /root/.cache/bazel/_bazel_root/bc57d4817a53cab8c785464da57d1983/execroot/ovms/bazel-out/k8-opt/bin/src/python/capi_python/ovmspybind.so src/python/capi_python/

python3 src/python/capi_python/ovms-pybind.py 

bazel --output_base=$PWD/output build //src/python/capi_python:ovmspybind.so

/model_serve

/root/.cache/bazel/_bazel_root/a249bbdc9c7f020e8d44a19d57754c73/execroot/ovms/bazel-out/k8-opt/bin/src/capi_python/ovmspybind.so
