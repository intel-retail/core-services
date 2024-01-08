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
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

// TestDockerStartContainer: test starting a container from the configuration yaml
func TestDockerStartContainer(t *testing.T) {
	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		require.NoError(t, err)
	}
	defer cli.Close()

	tests := []struct {
		name               string
		expectedErr        bool
		expectedContainers Containers
	}{
		{"valid container launch", false, CreateTestContainers("", "")},
		{"invalid container image", true, CreateTestContainersInvalidImage("", "")},
		{"invalid duplicate container names", true, CreateTestContainersDuplicates("", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			err := tt.expectedContainers.DockerStartContainer(ctx, cli)
			if err != nil {
				hasError = true
			}

			require.Equal(t, tt.expectedErr, hasError)

			containerList, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
			if err != nil {
				require.NoError(t, err)
			}

			for _, cont := range tt.expectedContainers.Containers {
				found := false
				for _, container := range containerList {
					if strings.Contains(container.Names[0], cont.Name) {
						found = true
						break
					}
				}
				require.True(t, found)
			}
			// cleanup
			for _, stopCont := range containerList {
				cli.ContainerRemove(ctx, stopCont.ID, types.ContainerRemoveOptions{})
			}
		})
	}
}

// TestSetHostNetwork: test loading the config yaml file
func TestSetHostNetwork(t *testing.T) {
	tests := []struct {
		name        string
		configDir   string
		networkMode container.NetworkMode
		ipcMode     container.IpcMode
	}{
		{"valid set host network", testConfigDir, container.NetworkMode("host"), container.IpcMode("host")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers("", "")
			tmpContainers.SetHostNetwork()

			for _, cont := range tmpContainers.Containers {
				require.Equal(t, cont.HostConfig.NetworkMode, tt.networkMode)
				require.Equal(t, cont.HostConfig.IpcMode, tt.ipcMode)
			}

		})
	}
}

// TestCreateVolumeMount: test creating volume mount structs
func TestCreateVolumeMount(t *testing.T) {
	// Get absolute path for test profile
	sourcePath, err := filepath.Abs("./test-profile/valid-profile")
	require.NoError(t, err)

	tests := []struct {
		name           string
		expectedErr    bool
		volume         string
		expectedVolume mount.Mount
	}{
		{"valid volume", false, "./test-profile/valid-profile:/test", mount.Mount{Type: mount.TypeBind, Source: sourcePath, Target: "/test", ReadOnly: false}},
		{"invalid volume format", true, "./test-profile/valid-profile", mount.Mount{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			volumeMount, err := CreateVolumeMount(tt.volume)
			if err != nil {
				hasError = true
			}

			require.Equal(t, tt.expectedErr, hasError)
			require.Equal(t, volumeMount, tt.expectedVolume)
		})
	}
}
