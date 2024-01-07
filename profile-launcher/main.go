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
	"flag"
	"fmt"

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

type envOverrideFlags []string

func main() {
	var envOverrides arrayFlags
	var volumes arrayFlags
	var configDir string
	var targetDevice string
	var inputSrc string
	flag.StringVar(&configDir, "configdir", "./test-profile", "Directory with the profile config")
	flag.StringVar(&targetDevice, "target_device", "", "Device you are targeting to run on. Default is CPU.")
	flag.StringVar(&inputSrc, "inputsrc", "", "Input for the profile to use.")
	flag.Var(&volumes, "v", "Volume mount for the container")
	flag.Var(&envOverrides, "e", "Environment overridees for the container")
	flag.Parse()

	containersArray, err := InitContainers(configDir, targetDevice, inputSrc, volumes, envOverrides)
	if err != nil {
		fmt.Errorf("Failed to run contaienrs %v", err)

	}

	if runErr := RunContainers(containersArray); runErr != nil {
		fmt.Errorf("Failed to run contaienrs %v", runErr)
	}
	return
}

func InitContainers(configDir string, targetDevice string, inputSrc string, volumes []string, envOverrides []string) (functions.Containers, error) {
	// Load yaml config
	containersArray, yamlErr := functions.GetYamlConfig(configDir)
	if yamlErr != nil {
		return functions.Containers{}, fmt.Errorf("Failed to load yaml config %v", yamlErr)
	}
	// Load ENV from .env file
	if err := containersArray.GetEnv(configDir); err != nil {
		return functions.Containers{}, fmt.Errorf("Failed to load ENV file %v", err)
	}
	containersArray.SetHostNetwork()

	// Set ENV overrides if any exist
	if len(envOverrides) > 0 {
		fmt.Println("Override Env")
		if err := containersArray.OverrideEnv(envOverrides); err != nil {
			return functions.Containers{}, err
		}
	}

	// Set Volumes
	if err := containersArray.SetVolumes(volumes); err != nil {
		return functions.Containers{}, err
	}

	// Set the target device ENV
	containersArray.TargetDevice = targetDevice
	if err := containersArray.SetTargetDevice(); err != nil {
		return functions.Containers{}, err
	}

	// Set the input source
	containersArray.InputSrc = inputSrc
	if err := containersArray.SetInputSrc(); err != nil {
		return functions.Containers{}, err
	}

	return containersArray, nil
}

func RunContainers(containersArray functions.Containers) error {
	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	// Run each container found in config
	if err := containersArray.DockerStartContainer(ctx, cli); err != nil {
		return err
	}
	return nil
}
