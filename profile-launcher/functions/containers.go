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

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"gopkg.in/yaml.v3"
)

func GetYamlConfig(configDir string) (Containers, error) {
	profileConfigPath := filepath.Join(configDir, "profile_config.yaml")
	contents, err := os.ReadFile(profileConfigPath)
	if err != nil {
		return Containers{}, fmt.Errorf("Unable to read config file: %v, error: %v",
			configDir, err)
	}

	containersArray := Containers{}
	err = yaml.Unmarshal(contents, &containersArray)
	if err != nil {
		return Containers{}, fmt.Errorf("error: %v", err)
	}

	return containersArray, nil
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
			if override != "" {
				overrideArray := strings.Split(override, "=")
				if len(overrideArray) != 2 {
					return fmt.Errorf("env format incorrect, ensure env is EnvName=Value.")
				}
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

// Setup volume mounts for the countainer
func (containerArray *Containers) SetVolumes(volumes []string) error {
	// First create the mounts from the volume inputs
	var volumeMountParam []mount.Mount
	for _, vol := range volumes {
		tmpVol, err := CreateVolumeMount(vol)
		if err != nil {
			return fmt.Errorf("Failed to create volume mount")
		}
		volumeMountParam = append(volumeMountParam, tmpVol)
	}

	// Second get the volumes found in the config yaml
	for contIndex, cont := range containerArray.Containers {
		for _, vol := range cont.Volumes {
			tmpVol, err := CreateVolumeMount(vol)
			if err != nil {
				return fmt.Errorf("Failed to create volume mount")
			}
			containerArray.Containers[contIndex].HostConfig.Mounts = append(containerArray.Containers[contIndex].HostConfig.Mounts, tmpVol)
		}
		// Append any input volumes to the config yaml volume list
		containerArray.Containers[contIndex].HostConfig.Mounts = append(containerArray.Containers[contIndex].HostConfig.Mounts, volumeMountParam...)
	}

	return nil
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
		TargetGpu := strconv.Itoa(TargetDeviceNum)
		// Set GPU Device
		containerArray.SetHostDevice("/dev/dri/renderD" + TargetGpu)
	} else {
		return fmt.Errorf("Target device not supported")
	}
	return nil
}

// Setup devices and other mounts based on the inputsrc
func (containerArray *Containers) SetInputSrc() error {
	if containerArray.InputSrc == "" {
		return errors.New("InputSrc was not set. Exiting profile launcher.")
	} else if strings.Contains(containerArray.InputSrc, "/video") {
		containerArray.SetHostDevice(containerArray.InputSrc)
	}

	for contIndex, _ := range containerArray.Containers {
		containerArray.Containers[contIndex].Envs = append(containerArray.Containers[contIndex].Envs, "INPUTSRC="+containerArray.InputSrc)
	}

	return nil
}
