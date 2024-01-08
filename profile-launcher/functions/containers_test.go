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
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"
)

// TestGetYamlConfig: test loading the config yaml file
func TestGetYamlConfig(t *testing.T) {
	tests := []struct {
		name               string
		configDir          string
		expectedErr        bool
		expectedContainers Containers
	}{
		{"valid profile config with 2 containers", testConfigDir, false, CreateTestContainers("", "")},
		{"invalid profile config", "./invalid", true, Containers{}},
		{"invalid profile format", testInvalidFormatDir, true, Containers{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			containersArray, err := GetYamlConfig(tt.configDir)
			if err != nil {
				hasError = true
			}

			require.Equal(t, tt.expectedContainers, containersArray)
			require.Equal(t, tt.expectedErr, hasError)
		})
	}
}

// GetEnv: test loading env file
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name            string
		configDir       string
		expectedErr     bool
		setTargetDevice string
	}{
		{"valid profile config with 2 containers", testConfigDir, false, ""},
		{"invalid profile config", "./invalid", true, ""},
		{"valid profile config with target device set", testConfigDir, false, "CPU"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers("", tt.setTargetDevice)

			hasError := false
			err := tmpContainers.GetEnv(tt.configDir)
			if err != nil {
				hasError = true
			}

			require.Equal(t, tt.expectedErr, hasError)
		})
	}
}

// OverrideEnv: test loading env file
func TestOverrideEnv(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr bool
		overrideEnv []string
		expectedEnv []string
	}{
		{"valid env overrides", false, []string{"TEST_ENV=test"}, []string{"TEST_ENV=test"}},
		{"valid new env", false, []string{"NEW_ENV=test"}, []string{"TEST_ENV=123", "NEW_ENV=test"}},
		{"invalid env overrides", true, []string{"TEST_ENV"}, []string{}},
		{"invalid new env", true, []string{"NEW_ENV"}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers("", "")
			for contIndex, _ := range tmpContainers.Containers {
				tmpContainers.Containers[contIndex].Envs = []string{"TEST_ENV=123"}
			}

			hasError := false
			err := tmpContainers.OverrideEnv(tt.overrideEnv)
			if err != nil {
				hasError = true
			}

			if tt.expectedErr != true {
				for _, cont := range tmpContainers.Containers {
					require.Equal(t, cont.Envs, tt.expectedEnv)
				}
			}
			require.Equal(t, tt.expectedErr, hasError)
		})
	}
}

// TestSetVolumes: test setting volume mounts
func TestSetVolumes(t *testing.T) {
	// Create test volume mounts
	volumeMount1, err := CreateVolumeMount("volume:volume")
	require.NoError(t, err)
	volumeMount2, err := CreateVolumeMount("test:test")
	require.NoError(t, err)

	tests := []struct {
		name               string
		expectedErr        bool
		volume             []string
		expectedVolumes    []mount.Mount
		expectedContainers Containers
	}{
		{"valid with no input volumes", false, []string{}, []mount.Mount{volumeMount1}, Containers{Containers: []Container{{Volumes: []string{"volume:volume"}}}}},
		{"valid with input volumes", false, []string{"test:test"}, []mount.Mount{volumeMount1, volumeMount2}, Containers{Containers: []Container{{Volumes: []string{"volume:volume"}}}}},
		{"invalid volume in config", true, []string{}, []mount.Mount{}, Containers{Containers: []Container{{Volumes: []string{"test"}}}}},
		{"invalid volume param", true, []string{"test"}, []mount.Mount{}, CreateTestContainers("", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			err := tt.expectedContainers.SetVolumes(tt.volume)
			if err != nil {
				hasError = true
			}

			if tt.expectedErr == false {
				for _, cont := range tt.expectedContainers.Containers {
					require.Equal(t, cont.HostConfig.Mounts, tt.expectedVolumes)
				}
			}
			require.Equal(t, tt.expectedErr, hasError)
		})
	}
}

// SetTargetDevice: test setting target device
func TestSetTargetDevice(t *testing.T) {
	tests := []struct {
		name            string
		expectedErr     bool
		setTargetDevice string
		isPrivileged    bool
		setHostDevice   []container.DeviceMapping
		hasDevices      bool
	}{
		{"valid no target device", false, "", true, []container.DeviceMapping{}, false},
		{"valid CPU target device", false, "CPU", false, []container.DeviceMapping{}, false},
		{"valid GPU target device", false, "GPU", true, []container.DeviceMapping{}, false},
		{"valid GPU.0 target device", false, "GPU.0", false, []container.DeviceMapping{{PathOnHost: "/dev/dri/renderD128", PathInContainer: "/dev/dri/renderD128", CgroupPermissions: "rwm"}}, true},
		{"invalid target device", true, "invalid", false, []container.DeviceMapping{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers("", tt.setTargetDevice)
			hasError := false
			err := tmpContainers.SetTargetDevice()
			if err != nil {
				hasError = true
			}
			require.Equal(t, tt.expectedErr, hasError)
			for _, cont := range tmpContainers.Containers {
				require.Equal(t, cont.HostConfig.Privileged, tt.isPrivileged)
				if tt.hasDevices {
					require.Equal(t, cont.HostConfig.Devices, tt.setHostDevice)

				}
			}
		})
	}
}

// TestSetInputSrc: test setting input src
func TestSetInputSrc(t *testing.T) {
	tests := []struct {
		name          string
		expectedErr   bool
		setInputSrc   string
		setHostDevice []container.DeviceMapping
		hasDevices    bool
	}{
		{"valid no input src", true, "", []container.DeviceMapping{}, false},
		{"valid USB input src", false, "/dev/video0", []container.DeviceMapping{{PathOnHost: "/dev/video0", PathInContainer: "/dev/video0", CgroupPermissions: "rwm"}}, true},
		{"valid RTSP input src", false, "RTSP://127.0.0.1:8554/camera_0", []container.DeviceMapping{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers(tt.setInputSrc, "")
			hasError := false
			err := tmpContainers.SetInputSrc()
			if err != nil {
				hasError = true
			}
			require.Equal(t, tt.expectedErr, hasError)
			for _, cont := range tmpContainers.Containers {
				if tt.hasDevices {
					require.Equal(t, cont.HostConfig.Devices, tt.setHostDevice)

				}
			}
		})
	}
}
