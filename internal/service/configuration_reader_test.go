/*
 Copyright (c) 2021-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package service_test

import (
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/domain"
	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/stretchr/testify/assert"
)

func Test_ConfigurationReader(t *testing.T) {
	type checkFn func(*testing.T, []domain.ArrayConnectionData, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, _ []domain.ArrayConnectionData, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []domain.ArrayConnectionData) func(t *testing.T, result []domain.ArrayConnectionData, err error) {
		return func(t *testing.T, result []domain.ArrayConnectionData, _ error) {
			assert.Equal(t, expectedOutput, result)
		}
	}

	hasError := func(t *testing.T, _ []domain.ArrayConnectionData, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (service.ConfigurationReader, string, []checkFn){
		"success json with no default system in config": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-with-no-default.json"
			configReader := service.ConfigurationReader{}

			expectedResult := []domain.ArrayConnectionData{
				{
					Username:  "admin",
					Password:  "password",
					SystemID:  "ID1",
					Endpoint:  "http://127.0.0.1",
					Insecure:  true,
					IsDefault: false,
				},
				{
					Username: "admin",
					Password: "password",
					SystemID: "ID2",
					Endpoint: "https://127.0.0.2",
					Insecure: true,
				},
			}

			return configReader, file, check(hasNoError, checkExpectedOutput(expectedResult))
		},
		"success yaml with no default system in config": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-with-no-default.yaml"
			configReader := service.ConfigurationReader{}

			expectedResult := []domain.ArrayConnectionData{
				{
					Username:  "admin",
					Password:  "password",
					SystemID:  "ID1",
					Endpoint:  "http://127.0.0.1",
					Insecure:  true,
					IsDefault: false,
				},
				{
					Username: "admin",
					Password: "password",
					SystemID: "ID2",
					Endpoint: "https://127.0.0.2",
					Insecure: true,
				},
			}

			return configReader, file, check(hasNoError, checkExpectedOutput(expectedResult))
		},
		"success yaml with skipCertificateValidation": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-with-skipCertificateValidation.yaml"
			configReader := service.ConfigurationReader{}

			expectedResult := []domain.ArrayConnectionData{
				{
					Username:                  "admin",
					Password:                  "password",
					SystemID:                  "ID1",
					Endpoint:                  "http://127.0.0.1",
					Insecure:                  false,
					SkipCertificateValidation: true,
					IsDefault:                 false,
				},
				{
					Username:                  "admin",
					Password:                  "password",
					SystemID:                  "ID2",
					Endpoint:                  "https://127.0.0.2",
					Insecure:                  false,
					SkipCertificateValidation: false,
				},
			}

			return configReader, file, check(hasNoError, checkExpectedOutput(expectedResult))
		},
		"error when file doesn't exist": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/non-existant-file.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has 0 storage sysytems": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-with-0-storage-systems.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file is empty": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-empty-file.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has invalid format": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-invalid-format.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has missing endpoint": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-missing-endpoint.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has missing password": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-missing-password.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has missing systemid": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-missing-systemid.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has missing username": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-missing-username.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when file has invalid default system": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/config-with-invalid-default-system.json"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
		"error when using directory as config file": func(*testing.T) (service.ConfigurationReader, string, []checkFn) {
			file := "testdata/"
			configReader := service.ConfigurationReader{}
			return configReader, file, check(hasError)
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			configReader, file, checkFns := tc(t)
			storageSystemConfiguration, err := configReader.GetStorageSystemConfiguration(file)
			for _, checkFn := range checkFns {
				checkFn(t, storageSystemConfiguration, err)
			}
		})
	}
}
