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

var testConfigDir = "../test-profile/valid-profile"
var testInvalidConfigDir = "../test-profile/invalid-test-profile"
var testInvalidFormatDir = "../test-profile/invalid-format-profile"

func CreateTestContainers(inputSrc string, targetDevice string) Containers {
	return Containers{
		InputSrc:     inputSrc,
		TargetDevice: targetDevice,
		Containers: []Container{{
			Name:                     "Client",
			DockerImage:              "test:dev",
			Entrypoint:               "/script/entrypoint.sh",
			EnvironmentVariableFiles: "profile.env",
			Volumes:                  []string{"./test-profile:/test-profile"},
		},
			{
				Name:                     "Server",
				DockerImage:              "test:dev",
				Entrypoint:               "/script/entrypoint2.sh",
				EnvironmentVariableFiles: "profile2.env",
				Volumes:                  []string{"./test-profile:/test-profile"},
			}},
	}
}

func CreateTestContainersInvalidImage(inputSrc string, targetDevice string) Containers {
	return Containers{
		InputSrc:     inputSrc,
		TargetDevice: targetDevice,
		Containers: []Container{{
			Name:                     "Client",
			DockerImage:              "",
			Entrypoint:               "/script/entrypoint.sh",
			EnvironmentVariableFiles: "profile.env",
			Volumes:                  []string{"./test-profile:/test-profile"},
		}},
	}
}

func CreateTestContainersDuplicates(inputSrc string, targetDevice string) Containers {
	return Containers{
		InputSrc:     inputSrc,
		TargetDevice: targetDevice,
		Containers: []Container{{
			Name:                     "Client",
			DockerImage:              "test:dev",
			Entrypoint:               "/script/entrypoint.sh",
			EnvironmentVariableFiles: "profile.env",
			Volumes:                  []string{"./test-profile:/test-profile"},
		},
			{
				Name:                     "Client",
				DockerImage:              "test:dev",
				Entrypoint:               "/script/entrypoint.sh",
				EnvironmentVariableFiles: "profile.env",
				Volumes:                  []string{"./test-profile:/test-profile"},
			}},
	}
}
