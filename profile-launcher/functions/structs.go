// ----------------------------------------------------------------------------------
// Copyright 2024 Intel Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	   http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
//
// ----------------------------------------------------------------------------------

package functions

import "github.com/docker/docker/api/types/container"

type Containers struct {
	Containers   []Container `yaml:"Containers"`
	InputSrc     string      `yaml:"InputSrc"`
	TargetDevice string      `yaml:"TargetDevice"`
}

type Container struct {
	Name                     string               `yaml:"Name"`
	DockerImage              string               `yaml:"DockerImage"`
	EnvironmentVariableFiles string               `yaml:"EnvironmentVariableFiles"`
	Envs                     []string             `yaml:"Envs"`
	Volumes                  []string             `yaml:"Volumes"`
	Entrypoint               string               `yaml:"Entrypoint"`
	HostConfig               container.HostConfig `yaml:"HostConfig"`
}
