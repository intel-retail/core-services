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
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

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
