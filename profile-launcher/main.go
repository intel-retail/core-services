// ----------------------------------------------------------------------------------
// Copyright 2023 Intel Corp.
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

package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Containers struct {
	Containers []Container `yaml:"Containers"`
}

type Container struct {
	Name                     string   `yaml:"Name"`
	DockerImage              string   `yaml:"DockerImage"`
	EnvironmentVariableFiles string   `yaml:"EnvironmentVariableFiles"`
	Volumes                  []string `yaml:"Volumes"`
}

// docker run --network host --user root --ipc=host \
// --name "$containerNameInstance" \
// --env-file "$DOT_ENV_FILE" \
// -e CONTAINER_NAME="$containerNameInstance" \
// $TARGET_USB_DEVICE \
// $TARGET_GPU_DEVICE \
// $volFullExpand \
// "$DOCKER_IMAGE" \
// bash -c '$DOCKER_CMD'

func main() {
	configDir := "/home/intel/projects/intel-retail/core-services/profile-launcher/test-profile/test_conf.yaml"
	GetYamlConfig(configDir)
}

func GetYamlConfig(configDir string) {
	contents, err := os.ReadFile(configDir)
	if err != nil {
		err = fmt.Errorf("Unable to read config file: %v, error: %v",
			configDir, err)
	}

	containersArray := Containers{}
	err = yaml.Unmarshal(contents, &containersArray)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(contents)
	fmt.Printf("%+v\n", containersArray)
}

func DockerStart() {
	fmt.Println("Starting Docker Container")
}
