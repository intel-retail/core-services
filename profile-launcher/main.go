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
	var volumes string
	var envOverrides string
	flag.StringVar(&configDir, "configdir", "./test-profile", "Directory with the profile config")
	flag.StringVar(&volumes, "v", "", "Volume mount for the container")
	flag.StringVar(&envOverrides, "e", "", "Environment overridees for the container")
	flag.Parse()
	fmt.Println(volumes)
	fmt.Println(envOverrides)

	containersArray := GetYamlConfig(configDir)
	envMap := GetEnv(configDir)

	envArray := OverrideEnv(envMap, envOverrides)
	for _, env := range envMap {
		fmt.Println("env:")
		fmt.Println(env)
	}
	fmt.Println(envArray)

	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Run each container found in config
	for _, cont := range containersArray.Containers {
		DockerStartContainer(cont, ctx, cli, envArray)
	}
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

func GetEnv(configDir string) map[string]string {
	profileConfigPath := filepath.Join(configDir, "profile.env")
	contents, err := os.ReadFile(profileConfigPath)
	if err != nil {
		err = fmt.Errorf("Unable to read config file: %v, error: %v",
			configDir, err)
	}

	splitEnv := strings.Split(string(contents[:]), "\n")
	envArray := make(map[string]string)
	for _, env := range splitEnv {
		envs := strings.Split(string(env[:]), "=")
		envArray[envs[0]] = envs[1]
	}
	return envArray
}

func OverrideEnv(envMap map[string]string, envOverrides string) []string {
	if envOverrides == "" {
		return []string{}
	}

	splitEnvOverride := strings.Split(string(envOverrides[:]), ",")
	for _, env := range splitEnvOverride {
		envs := strings.Split(string(env[:]), "=")
		envMap[envs[0]] = envs[1]
	}

	var envArray []string
	for key, value := range envMap {
		envArray = append(envArray, key+"="+value)
	}
	return envArray
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
