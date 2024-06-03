//*****************************************************************************
// Copyright 2024 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

#include "src/capi_frontend/server_settings.hpp"
#include "src/status.hpp"
#include "src/server.hpp"

#include <pybind11/pybind11.h>
#include <pybind11/stl.h>

namespace py = pybind11;
using namespace ovms;

ovms::ServerSettingsImpl ServerSettingsNew(){
    ovms::ServerSettingsImpl serverSettings;
    serverSettings.grpcPort = 9178;
    serverSettings.restPort = 0;
    serverSettings.grpcWorkers = 1;
    serverSettings.grpcBindAddress = "0.0.0.0";
    serverSettings.restBindAddress = "0.0.0.0";
    serverSettings.metricsEnabled = false;
    serverSettings.logLevel = "INFO";
    return serverSettings;
}

ovms::ModelsSettingsImpl ModelSettingsNew(){
    ovms::ModelsSettingsImpl modelSettings;
    modelSettings.configPath = "/tmp/config.yml";
    return modelSettings;
}

PYBIND11_MODULE(ovmspybind, m) {
    py::class_<ServerSettingsImpl>(m, "ServerSettingsImpl")
        .def(py::init<>()) // <-- bind the default constructor
        .def_readwrite("grpcPort", &ServerSettingsImpl::grpcPort);
    m.def("ServerSettingsNew", &ServerSettingsNew, "New server settings function");

    py::class_<ModelsSettingsImpl>(m, "ModelsSettingsImpl")
        .def(py::init<>()) // <-- bind the default constructor
        .def_readwrite("configPath", &ModelsSettingsImpl::configPath);
    m.def("ModelSettingsNew", &ModelSettingsNew, "New server settings function");

    py::enum_<StatusCode>(m, "StatusCode")
        .value("OK", ovms::StatusCode::OK);

    py::class_<Status>(m, "Status")
        .def("getCode", &Status::getCode);

    py::class_<Server>(m, "Server")
        .def_static("instance", &Server::instance, py::return_value_policy::reference)
        .def("start", py::overload_cast<ServerSettingsImpl*, ModelsSettingsImpl*>(&Server::start))
        .def("isLive", &Server::isLive);
}