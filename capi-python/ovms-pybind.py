import sys
# caution: path[0] is reserved for script path (or '' in REPL)
sys.path.insert(1, './')

import ovmspybind
import time

serverSet = ovmspybind.ServerSettingsNew()
print(serverSet.grpcPort)
serverSet.grpcPort = 8080
print(serverSet.grpcPort)

modelSet = ovmspybind.ModelSettingsNew()
print(modelSet.configPath)
modelSet.configPath = "config.json"
print(modelSet.configPath)

status = ovmspybind.StatusCode.OK
print(status)
print(status.value)

# server = ovmspybind.Server
server = ovmspybind.Server.instance()
print(server)
print(server.isLive())
startStatus = server.start(serverSet, modelSet)
print(startStatus.getCode())

while True:
    print(server.isLive())
    time.sleep(5)

