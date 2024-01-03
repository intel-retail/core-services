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
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

// A TEST RUN STRING
// go run main.go -e TEST_ENV=aaa,NEW=abc

type Containers struct {
	Containers []Container `yaml:"Containers"`
}

type Container struct {
	Name                     string   `yaml:"Name"`
	DockerImage              string   `yaml:"DockerImage"`
	EnvironmentVariableFiles string   `yaml:"EnvironmentVariableFiles"`
	Envs                     []string `yaml:"Envs"`
	Volumes                  []string `yaml:"Volumes"`
	InputSrc                 string   `yaml:"InputSrc"`
	Entrypoint               string   `yaml:"Entrypoint"`
}

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

// docker run --network host --user root --ipc=host \
// $TARGET_USB_DEVICE \
// $TARGET_GPU_DEVICE \

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

	containersArray := GetYamlConfig(configDir)
	if err := containersArray.GetEnv(configDir); err != nil {
		os.Exit(-1)
	}

	if len(envOverrides) > 0 {
		fmt.Println("Override Env")
		if err := containersArray.OverrideEnv(envOverrides); err != nil {
			os.Exit(-1)
		}
	}

	// Set the target device ENV
	containersArray.SetTargetDevice()

	if inputSrc == "" {
		fmt.Errorf("InputSrc was not set. Exiting profile launcher.")
		os.Exit(-1)
	}

	// // Setup Docker CLI
	// ctx := context.Background()
	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	panic(err)
	// }
	// defer cli.Close()

	// // Run each container found in config
	// 	DockerStartContainer(cont, ctx, cli, envArray)
}

func GetYamlConfig(configDir string) Containers {
	profileConfigPath := filepath.Join(configDir, "profile_config.yaml")
	contents, err := os.ReadFile(profileConfigPath)
	if err != nil {
		err = fmt.Errorf("Unable to read config file: %v, error: %v",
			configDir, err)
	}

	containersArray := Containers{}
	err = yaml.Unmarshal(contents, &containersArray)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return containersArray
}

func (containerArray *Containers) GetEnv(configDir string) error {
	for i, cont := range containerArray.Containers {
		profileConfigPath := filepath.Join(configDir, cont.EnvironmentVariableFiles)
		contents, err := os.ReadFile(profileConfigPath)
		if err != nil {
			err = fmt.Errorf("Unable to read config file: %v, error: %v",
				configDir, err)
			return err
		}
		containerArray.Containers[i].Envs = strings.Split(string(contents[:]), "\n")
	}
	return nil
}

func (containerArray *Containers) OverrideEnv(envOverrides []string) error {
	for contIndex, cont := range containerArray.Containers {
		for _, override := range envOverrides {
			notFound := true
			overrideArray := strings.Split(override, "=")
			if override != "" && len(overrideArray) == 2 {
				for envIndex, env := range cont.Envs {
					if strings.Contains(env, overrideArray[0]+"=") {
						containerArray.Containers[contIndex].Envs[envIndex] = override
						notFound = false
					}
				}
				if notFound {
					containerArray.Containers[contIndex].Envs = append(containerArray.Containers[contIndex].Envs, override)
				}
			}
		}
	}
	return nil
}

func (con *Containers) SetTargetDevice() {

}

func (containerArray *Containers) DockerStartContainer(ctx context.Context, cli *client.Client) {
	for _, cont := range containerArray.Containers {
		fmt.Println("Starting Docker Container")
		fmt.Printf("%+v\n", cont)

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:      cont.DockerImage,
			Env:        cont.Envs,
			Entrypoint: []string{cont.Entrypoint},
		},
			&container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeBind,
						Source: "/home/intel/projects/intel-retail/core-services/profile-launcher/test-profile",
						Target: "/test-profile",
					},
				},
			},
			nil, nil, cont.Name)
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
	}
}
