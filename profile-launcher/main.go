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
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

// A TEST RUN STRING
// go run main.go -e TEST_ENV=aaa,NEW=abc --configdir /use-cases/dlstreamer/res --inputsrc /dev/video4 --target_device CPU

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
	containersArray := GetYamlConfig(configDir)
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
	containersArray.SetInputSrc()

	if inputSrc == "" {
		fmt.Errorf("InputSrc was not set. Exiting profile launcher.")
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
		// Set Target Device ENV
		if containerArray.TargetDevice != "" {
			containerArray.Containers[i].Envs = append(containerArray.Containers[i].Envs, containerArray.TargetDevice)
		}
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

func (containerArray *Containers) SetVolumes(volumes []string) error {
	var volumeMountParam []mount.Mount
	for _, vol := range volumes {
		tmpVol, err := CreateVolumeMount(vol)
		if err != nil {
			return fmt.Errorf("Failed to create volume mount")
		}
		volumeMountParam = append(volumeMountParam, tmpVol)
	}

	for contIndex, cont := range containerArray.Containers {
		for _, vol := range cont.Volumes {
			tmpVol, err := CreateVolumeMount(vol)
			if err != nil {
				return fmt.Errorf("Failed to create volume mount")
			}
			containerArray.Containers[contIndex].HostConfig.Mounts = append(containerArray.Containers[contIndex].HostConfig.Mounts, tmpVol)
		}
		containerArray.Containers[contIndex].HostConfig.Mounts = append(containerArray.Containers[contIndex].HostConfig.Mounts, volumeMountParam...)
	}

	return nil
}

func CreateVolumeMount(vol string) (mount.Mount, error) {
	volSplit := strings.Split(vol, ":")
	sourcePath, err := filepath.Abs(volSplit[0])
	if err != nil {
		return mount.Mount{}, fmt.Errorf("Failed to get volume path %v", err)
	}

	return mount.Mount{
		Type:     mount.TypeBind,
		Source:   sourcePath,
		Target:   volSplit[1],
		ReadOnly: false,
	}, nil
}

// Setup device mounts and ENV based on the targe device input
func (containerArray *Containers) SetTargetDevice() error {
	if containerArray.TargetDevice == "" {
		containerArray.SetPrivileged()
	} else if containerArray.TargetDevice == "CPU" {
		// CPU do nothing
		return nil
	} else if containerArray.TargetDevice == "GPU" ||
		strings.Contains(containerArray.TargetDevice, "MULTI") ||
		containerArray.TargetDevice == "AUTO" {
		// GPU set to privileged so we can access all GPU
		containerArray.SetPrivileged()
	} else if strings.Contains(containerArray.TargetDevice, "GPU.") {
		// GPU.X set access to a specific GPU device and set the correct ENV
		TargetDeviceArray := strings.Split(containerArray.TargetDevice, ".")
		DeviceNum, errNum := strconv.Atoi(TargetDeviceArray[1])
		if errNum != nil {
			return errNum
		}
		TargetDeviceNum := 128 + DeviceNum
		TargetGpu := "GPU." + strconv.Itoa(TargetDeviceNum)
		// Set GPU Device
		containerArray.SetHostDevice("/dev/dri/renderD\"" + TargetGpu)
	} else {
		return fmt.Errorf("Target device not supported")
	}
	return nil
}

// Set the container to privileged mode
func (containerArray *Containers) SetPrivileged() {
	for contIndex, _ := range containerArray.Containers {
		containerArray.Containers[contIndex].HostConfig.Privileged = true
	}
}

// Set container to use host network
func (containerArray *Containers) SetHostNetwork() {
	for contIndex, _ := range containerArray.Containers {
		containerArray.Containers[contIndex].HostConfig.NetworkMode = "host"
		containerArray.Containers[contIndex].HostConfig.IpcMode = "host"
	}
}

// Setup the device mount
func (containerArray *Containers) SetHostDevice(device string) {
	deviceMount := container.DeviceMapping{
		PathOnHost:        device,
		PathInContainer:   device,
		CgroupPermissions: "rwm",
	}

	for contIndex, _ := range containerArray.Containers {
		containerArray.Containers[contIndex].HostConfig.Devices = append(containerArray.Containers[contIndex].HostConfig.Devices, deviceMount)
	}
}

// Setup devices and other mounts based on the inputsrc
func (containerArray *Containers) SetInputSrc() {
	if strings.Contains(containerArray.InputSrc, "/video") {
		containerArray.SetHostDevice(containerArray.InputSrc)
	}

	for contIndex, _ := range containerArray.Containers {
		containerArray.Containers[contIndex].Envs = append(containerArray.Containers[contIndex].Envs, "INPUTSRC="+containerArray.InputSrc)
	}
}

// Create and start the Docker container
func (containerArray *Containers) DockerStartContainer(ctx context.Context, cli *client.Client) {
	for _, cont := range containerArray.Containers {
		fmt.Println("Starting Docker Container")
		fmt.Printf("%+v\n", cont)

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image:      cont.DockerImage,
			Env:        cont.Envs,
			Entrypoint: strings.Split(cont.Entrypoint, " "),
		},
			&cont.HostConfig,
			nil, nil, cont.Name)
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
	}
}
