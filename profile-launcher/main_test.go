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
		},
			{
				Name:                     "Server",
				DockerImage:              "test:dev",
				Entrypoint:               "/script/entrypoint2.sh",
				EnvironmentVariableFiles: "profile2.env",
				Volumes:                  []string{"./test-profile:/test-profile"},
				Envs:                     []string{"NEW_ENV=456", "NEW_2ENV=efg", "INPUTSRC=/dev/video0"},
			}},
	}
}

func TestMain(t *testing.T) {
	// Setup Docker CLI
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		require.NoError(t, err)
	}
	defer cli.Close()

	tests := []struct {
		name               string
		args               []string
		expectedContainers functions.Containers
	}{
		// {"valid container launch", []string{"--configdir", "./test-profile", "--inputsrc", "/dev/video0", "--target_device", "CPU", "-e", "test=123", "-v", "./test-profile:/test"}, CreateTestContainers("", "")},
		// {"invalid container init", []string{"--configdir", "./test-profile/invalid-format-profile"}, functions.Containers{}},
		{"invalid container run", []string{"--configdir", "./test-profile/invalid-image-profile", "--inputsrc", "/dev/video0"}, functions.Containers{Containers: []functions.Container{{Name: "Client"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append(os.Args, tt.args...)
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

func MainTestRun() {

}

// TestInitContainers: test creating thhasError := false container struct
func TestInitContainers(t *testing.T) {
	tests := []struct {
		name                   string
		configDir              string
		targetDevice           string
		inputSrc               string
		volumes                []string
		envOverrides           []string
		expectedErr            bool
		expectedContainers     functions.Containers
		expectedContainerCount int
	}{
		{"valid with only flags", "./test-profile", "CPU", "/dev/video0", []string{}, []string{}, false, CreateTestContainers("CPU", "/dev/video0"), 2},
		{"valid with env overrides", "./test-profile", "CPU", "/dev/video0", []string{}, []string{"TEST_ENV=def"}, false, CreateTestContainers("CPU", "/dev/video0"), 2},
		{"valid with volume input", "./test-profile", "CPU", "/dev/video0", []string{"./test-profile:/test-profile"}, []string{}, false, CreateTestContainers("CPU", "/dev/video0"), 2},
		{"invalid no configdir set", "", "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid config format", "./test-profile/invalid-test-profile", "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid miussing env", "./test-profile/invalid-missing-env", "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid inputSrc not set", "./test-profile", "", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid targetDevice not set", "./test-profile", "Fake", "", []string{}, []string{}, true, functions.Containers{}, 0},
		{"invalid with env overrides", "./test-profile", "CPU", "/dev/video0", []string{}, []string{"TEST_ENV"}, true, CreateTestContainers("CPU", "/dev/video0"), 0},
		{"invalid with volume input", "./test-profile", "CPU", "/dev/video0", []string{"./test-profile"}, []string{}, true, CreateTestContainers("CPU", "/dev/video0"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
