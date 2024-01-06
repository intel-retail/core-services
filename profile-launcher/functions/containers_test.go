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

var testConfigDir = "../test-profile"

func CreateTestContainers(inputSrc string, targetDevice string) Containers {
	return Containers{
		InputSrc:     inputSrc,
		TargetDevice: targetDevice,
		Containers: []Container{{
			Name:                     "OVMSClient",
			DockerImage:              "test:dev",
			Entrypoint:               "/script/entrypoint.sh",
			EnvironmentVariableFiles: "profile.env",
			Volumes:                  []string{"./test-profile:/test-profile"},
		},
			{
				Name:                     "OVMSServer",
				DockerImage:              "test:dev",
				Entrypoint:               "/script/entrypoint2.sh",
				EnvironmentVariableFiles: "profile2.env",
				Volumes:                  []string{"./test-profile:/test-profile"},
			}},
	}
}

// TestGetYamlConfig: test loading the config yaml file
func TestGetYamlConfig(t *testing.T) {
	tests := []struct {
		name               string
		configDir          string
		expectedContainers Containers
	}{
		{"valid profile config with 2 containers", testConfigDir, CreateTestContainers("", "")},
		{"invalid profile config", "./invalid", Containers{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containersArray := GetYamlConfig(tt.configDir)
			require.Equal(t, tt.expectedContainers, containersArray)
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

			for _, cont := range tmpContainers.Containers {
				require.Equal(t, cont.Envs, tt.expectedEnv)
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
		name            string
		expectedErr     bool
		volume          []string
		expectedVolumes []mount.Mount
	}{
		{"valid with no input volumes", false, []string{}, []mount.Mount{volumeMount1}},
		{"valid with input volumes", false, []string{"test:test"}, []mount.Mount{volumeMount1, volumeMount2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpContainers := CreateTestContainers("", "")
			for contIndex, _ := range tmpContainers.Containers {
				tmpContainers.Containers[contIndex].Volumes = []string{"volume:volume"}
			}

			hasError := false
			err := tmpContainers.SetVolumes(tt.volume)
			if err != nil {
				hasError = true
			}

			for _, cont := range tmpContainers.Containers {
				require.Equal(t, cont.HostConfig.Mounts, tt.expectedVolumes)
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
