/*
 Copyright (c) 2021-2023 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dell/karavi-metrics-powerflex/internal/domain"
	"sigs.k8s.io/yaml"
)

// ConfigurationReader handles reading of the storage system configuration secret
type ConfigurationReader struct{}

// GetStorageSystemConfiguration returns a storage system from the configuration file
// If no default system is supplied, the first system in the list is returned
func (c *ConfigurationReader) GetStorageSystemConfiguration(file string) ([]domain.ArrayConnectionData, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s", fmt.Sprintf("File %s does not exist", file))
	}

	config, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("%s", fmt.Sprintf("File %s errors: %v", file, err))
	}

	if string(config) == "" {
		return nil, fmt.Errorf("arrays details are not provided in vxflexos-config secret")
	}

	connectionData := make([]domain.ArrayConnectionData, 0)
	// support backward compatibility
	config, err = yaml.JSONToYAML(config)
	if err != nil {
		return nil, fmt.Errorf("%s", fmt.Sprintf("converting json to yaml: %v", err))
	}

	err = yaml.Unmarshal(config, &connectionData)
	if err != nil {
		return nil, fmt.Errorf("%s", fmt.Sprintf("Unable to parse the credentials: %v", err))
	}

	if len(connectionData) == 0 {
		return nil, fmt.Errorf("no arrays are provided in vxflexos-config secret")
	}

	for i, c := range connectionData {
		err := validateStorageSystem(c, i)
		if err != nil {
			return nil, err
		}
	}

	return connectionData, nil
}

func validateStorageSystem(system domain.ArrayConnectionData, i int) error {
	if system.SystemID == "" {
		return fmt.Errorf("%s", fmt.Sprintf("invalid value for system name at index %d", i))
	}
	if system.Username == "" {
		return fmt.Errorf("%s", fmt.Sprintf("invalid value for Username at index %d", i))
	}
	if system.Password == "" {
		return fmt.Errorf("%s", fmt.Sprintf("invalid value for Password at index %d", i))
	}
	if system.Endpoint == "" {
		return fmt.Errorf("%s", fmt.Sprintf("invalid value for Endpoint at index %d", i))
	}
	return nil
}
