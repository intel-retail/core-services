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
	Volumes                  []string `yaml:"Volumes"`
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
// --name "$containerNameInstance" \
// --env-file "$DOT_ENV_FILE" \
// -e CONTAINER_NAME="$containerNameInstance" \
// $TARGET_USB_DEVICE \
// $TARGET_GPU_DEVICE \
// $volFullExpand \
// "$DOCKER_IMAGE" \
// bash -c '$DOCKER_CMD'
type envOverrideFlags []string

func main() {
	var configDir string
	flag.StringVar(&configDir, "configdir", "./test-profile", "Directory with the profile config")
	flag.Var(&volumes, "v", "Volume mount for the container")
	flag.Var(&envOverrides, "e", "Environment overridees for the container")
	flag.Parse()

	containersArray := GetYamlConfig(configDir)
	envArray := GetEnv(configDir)
	fmt.Println(containersArray)
	fmt.Println(envArray)

	if len(envOverrides) > 0 {
		fmt.Println("Override Env")
		envArray = OverrideEnv(envArray, envOverrides)
		fmt.Println(envArray)
	}

	// // Setup Docker CLI
	// ctx := context.Background()
	// cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	panic(err)
	// }
	// defer cli.Close()

	// // Run each container found in config
	// for _, cont := range containersArray.Containers {
	// 	DockerStartContainer(cont, ctx, cli, envArray)
	// }
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

func GetEnv(configDir string) []string {
	profileConfigPath := filepath.Join(configDir, "profile.env")
	contents, err := os.ReadFile(profileConfigPath)
	if err != nil {
		err = fmt.Errorf("Unable to read config file: %v, error: %v",
			configDir, err)
	}

	return strings.Split(string(contents[:]), "\n")
}

func OverrideEnv(envArray []string, envOverrides []string) []string {
	tmpEnvArray := envArray
	for _, override := range envOverrides {
		notFound := true
		overrideArray := strings.Split(override, "=")
		if override != "" && len(overrideArray) == 2 {
			for i, env := range envArray {
				if strings.Contains(env, overrideArray[0]+"=") {
					tmpEnvArray[i] = override
					notFound = false
				}
			}
			if notFound {
				tmpEnvArray = append(tmpEnvArray, override)
			}
		}
	}
	return tmpEnvArray
}

func DockerStartContainer(cont Container, ctx context.Context, cli *client.Client, env []string) {
	fmt.Println("Starting Docker Container")
	fmt.Printf("%+v\n", cont)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      cont.DockerImage,
		Env:        env,
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
