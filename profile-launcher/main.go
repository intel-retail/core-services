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

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/docker/docker/client"
	"github.com/intel-retail/core-services/profile-launcher/functions"
)

// Array flag when variables can be used multiple times
type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var envOverrides arrayFlags
var volumes arrayFlags

type envOverrideFlags []string

func main() {
	var configDir string
	var targetDevice string
	var inputSrc string
	flag.StringVar(&configDir, "configdir", "./test-profile", "Directory with the profile config")
	flag.StringVar(&targetDevice, "target_device", "", "Device you are targeting to run on. Default is CPU.")
	flag.StringVar(&inputSrc, "inputsrc", "", "Input for the profile to use.")
	flag.Var(&volumes, "v", "Volume mount for the container")
	flag.Var(&envOverrides, "e", "Environment overridees for the container")
	flag.Parse()

	// Load yaml config
	containersArray := functions.GetYamlConfig(configDir)
	// Load ENV from .env file
	if err := containersArray.GetEnv(configDir); err != nil {
		fmt.Errorf("Failed to load ENV file %v", err)
		os.Exit(-1)
	}
	containersArray.SetHostNetwork()

	// Set ENV overrides if any exist
	if len(envOverrides) > 0 {
		fmt.Println("Override Env")
		if err := containersArray.OverrideEnv(envOverrides); err != nil {
			fmt.Errorf("Failed to over ride input ENV values %v", err)
			os.Exit(-1)
		}
	}

	// Set Volumes
	if err := containersArray.SetVolumes(volumes); err != nil {
		fmt.Errorf("Failed to load Volumes from config file %v", err)
		os.Exit(-1)
	}

	// Set the target device ENV
	containersArray.TargetDevice = targetDevice
	containersArray.SetTargetDevice()
	// Set the input source
	containersArray.InputSrc = inputSrc
	err := containersArray.SetInputSrc()
	if err != nil {
		fmt.Errorf("%v", err)
		os.Exit(-1)
	}

	containersArray2, _ := json.Marshal(containersArray)
	fmt.Println(string(containersArray2))

	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Run each container found in config
	containersArray.DockerStartContainer(ctx, cli)
}
