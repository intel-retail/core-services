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
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

// func (containerArray *Containers) DockerStartContainer(ctx context.Context, cli *client.Client) {

// TestSetHostNetwork: test loading the config yaml file
func TestSetHostNetwork(t *testing.T) {
	tests := []struct {
		name        string
		configDir   string
		networkMode container.NetworkMode
		ipcMode     container.IpcMode
	}{
		{"valid set host network", "./test-profile", container.NetworkMode("host"), container.IpcMode("host")},
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
	sourcePath, err := filepath.Abs("./test-profile")
	require.NoError(t, err)

	tests := []struct {
		name           string
		expectedErr    bool
		volume         string
		expectedVolume mount.Mount
	}{
		{"valid with no input volumes", false, "./test-profile:/test", mount.Mount{Type: mount.TypeBind, Source: sourcePath, Target: "/test", ReadOnly: false}},
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
