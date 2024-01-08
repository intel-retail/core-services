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
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/intel-retail/core-services/profile-launcher/functions"
	"github.com/stretchr/testify/require"
)

func CreateTestContainers(inputSrc string, targetDevice string) functions.Containers {
	return functions.Containers{
		InputSrc:     inputSrc,
		TargetDevice: targetDevice,
		Containers: []functions.Container{{
			Name:                     "Client",
			DockerImage:              "test:dev",
			Entrypoint:               "/script/entrypoint.sh",
			EnvironmentVariableFiles: "profile.env",
			Volumes:                  []string{"./test-profile:/test-profile"},
			Envs:                     []string{"TEST_ENV=123", "TEST_ENV2=abc", "INPUTSRC=/dev/video0"},
		}},
	}
}

const (
	testConfigDir = "./test-profile/main-test-profile"
	validYaml     = `Containers:
    - Name: Client
      DockerImage: test:dev
      EnvironmentVariableFiles: profile.env
      Entrypoint: /script/entrypoint.sh
      Volumes: 
        - ./test-profile:/test-profile`
	invalidImageYaml = `Containers:
    - Name: Client
      DockerImage: ""
      EnvironmentVariableFiles: profile.env
      Entrypoint: /script/entrypoint.sh
      Volumes: 
        - ./test-profile:/test-profile`
	duplicateNameYaml = `Containers:
    - Name: Client
      DockerImage: test:dev
      EnvironmentVariableFiles: profile.env
      Entrypoint: /script/entrypoint.sh
      Volumes: 
        - ./test-profile/invalid-test-profile:/test-profile
	- Name: Client
      DockerImage: test:dev
      EnvironmentVariableFiles: profile.env
      Entrypoint: /script/entrypoint.sh
      Volumes: 
        - ./test-profile/invalid-test-profile:/test-profile`
	invalidEnvFile = `Containers:
    - Name: Client
      DockerImage: ""
      EnvironmentVariableFiles: fake.env
      Entrypoint: /script/entrypoint.sh
      Volumes: 
        - ./test-profile:/test-profile`
	invalidConfig = `invalid`
)

func WriteYamlConfig(yamlString string) error {
	yamlByte := []byte(yamlString)
	err := os.WriteFile("./test-profile/main-test-profile/profile_config.yaml", yamlByte, 0644)
	if err != nil {
		return err
	}
	return nil
}

// TestMain: test the main function
func TestMain(t *testing.T) {
	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		require.NoError(t, err)
	}
	defer cli.Close()

	os.Args = append(os.Args, []string{"--configdir", "./test-profile/main-test-profile", "--inputsrc", "/dev/video0", "--target_device", "CPU", "-e", "test=123", "-v", "./test-profile:/test"}...)
	tests := []struct {
		name               string
		profile            string
		expectedContainers functions.Containers
	}{
		// {"valid container launch", validYaml, CreateTestContainers("", "")},
		// {"invalid container init", invalidImageYaml, functions.Containers{}},
		{"invalid container init", duplicateNameYaml, functions.Containers{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, WriteYamlConfig(tt.profile))
			main()

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

// TestInitContainers: test creating thhasError := false container struct
func TestInitContainers(t *testing.T) {
	tests := []struct {
		name                   string
		profile                string
		configDir              string
		targetDevice           string
		inputSrc               string
		volumes                []string
		envOverrides           []string
		expectedErr            bool
		expectedContainers     functions.Containers
		expectedContainerCount int
	}{
		{"valid with only flags", validYaml, testConfigDir, "CPU", "/dev/video0", []string{}, []string{}, false, CreateTestContainers("CPU", "/dev/video0"), 1},
		{"valid with env overrides", validYaml, testConfigDir, "CPU", "/dev/video0", []string{}, []string{"TEST_ENV=def"}, false, CreateTestContainers("CPU", "/dev/video0"), 1},
		{"valid with volume input", validYaml, testConfigDir, "CPU", "/dev/video0", []string{"./test-profile:/test-profile"}, []string{}, false, CreateTestContainers("CPU", "/dev/video0"), 1},
		{"invalid no configdir set", "", "", "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid config format", invalidConfig, testConfigDir, "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid missing env", invalidEnvFile, testConfigDir, "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid inputSrc not set", validYaml, testConfigDir, "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid targetDevice not set", validYaml, testConfigDir, "Fake", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid with env overrides", validYaml, testConfigDir, "CPU", "/dev/video0", []string{}, []string{"TEST_ENV"}, true, CreateTestContainers("CPU", "/dev/video0"), 0},
		{"invalid with volume input", validYaml, testConfigDir, "CPU", "/dev/video0", []string{testConfigDir}, []string{}, true, CreateTestContainers("CPU", "/dev/video0"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, WriteYamlConfig(tt.profile))
			tmpContainers := CreateTestContainers("", "")
			tmpContainers.SetHostNetwork()

			hasError := false
			containersArray, err := InitContainers(tt.configDir, tt.targetDevice, tt.inputSrc, tt.volumes, tt.envOverrides)
			if err != nil {
				hasError = true

			}
			require.Equal(t, tt.expectedErr, hasError)
			require.Equal(t, tt.expectedContainerCount, len(containersArray.Containers))
			// require.Equal(t, tt.expectedContainers, containersArray)
		})
	}
}

// TestRunContainers: test running the containers
func TestRunContainers(t *testing.T) {
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
		expectedContainers functions.Containers
	}{
		{"valid container launch", false, CreateTestContainers("", "")},
		{"invalid container", true, functions.Containers{Containers: []functions.Container{{Name: "invalidTest"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false
			err := RunContainers(tt.expectedContainers)
			if err != nil {
				hasError = true

			}
			require.Equal(t, tt.expectedErr, hasError)

			containerList, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
			if err != nil {
				require.NoError(t, err)
			}

			// cleanup
			for _, stopCont := range containerList {
				cli.ContainerRemove(ctx, stopCont.ID, types.ContainerRemoveOptions{})
			}
		})
	}
}
