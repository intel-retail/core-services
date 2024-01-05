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
	"testing"

	"github.com/stretchr/testify/require"
)

var TestContainers = Containers{
	InputSrc:     "",
	TargetDevice: "",
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

// TestGetYamlConfig: test loading the config yaml file
func TestGetYamlConfig(t *testing.T) {
	tests := []struct {
		name               string
		configDir          string
		expectedContainers Containers
	}{
		{"valid profile config with 2 containers", "./test-profile", TestContainers},
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
		name               string
		configDir          string
		expectedErr        bool
		expectedContainers Containers
		setTargetDevice    bool
	}{
		{"valid profile config with 2 containers", "./test-profile", false, TestContainers, false},
		{"invalid profile config", "./invalid", true, TestContainers, false},
		{"valid profile config with target device set", "./test-profile", false, TestContainers, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set target device
			if tt.setTargetDevice {
				tt.expectedContainers.TargetDevice = "CPU"
			}

			hasError := false
			err := tt.expectedContainers.GetEnv(tt.configDir)
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
		name               string
		expectedErr        bool
		expectedContainers Containers
		overrideEnv        []string
		expectedEnv        []string
	}{
		{"valid env overrides", false, TestContainers, []string{"TEST_ENV=test"}, []string{"TEST_ENV=test"}},
		{"valid new env", false, TestContainers, []string{"NEW_ENV=test"}, []string{"TEST_ENV=123", "NEW_ENV=test"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for contIndex, _ := range tt.expectedContainers.Containers {
				tt.expectedContainers.Containers[contIndex].Envs = []string{"TEST_ENV=123"}
			}

			hasError := false
			err := tt.expectedContainers.OverrideEnv(tt.overrideEnv)
			if err != nil {
				hasError = true
			}

			for _, cont := range tt.expectedContainers.Containers {
				require.Equal(t, cont.Envs, tt.expectedEnv)
			}
			require.Equal(t, tt.expectedErr, hasError)
		})
	}
}

// func (containerArray *Containers) SetVolumes(volumes []string) error {
// func CreateVolumeMount(vol string) (mount.Mount, error) {
// func (containerArray *Containers) SetTargetDevice() error {
// func (containerArray *Containers) SetPrivileged() {
// func (containerArray *Containers) SetHostNetwork() {
// func (containerArray *Containers) SetHostDevice(device string) {
// func (containerArray *Containers) SetInputSrc() {
// func (containerArray *Containers) DockerStartContainer(ctx context.Context, cli *client.Client) {
