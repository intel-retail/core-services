# Python binding for C APIs

The python C API bindings enable developers to directly access OVMS. This bypasses the GRPC protocol to gain a slight performance advantage. This is intended for applications running in a single container and machine. It is recommended to use the GRPC interface for distributed deployments.

## Building the shared library

1. make build ovms-build-image
2. make build-capi-python

## Running the example

1. make run
2. cd /tmp/project
3. cp /output/ovmspybind.so .
4. python3 ovms-pybind.py
5. See isLive status True in a loop
