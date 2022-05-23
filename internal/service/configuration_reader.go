// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ArrayConnectionData contains data required to connect to array
type ArrayConnectionData struct {
	SystemID  string `json:"systemID"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Endpoint  string `json:"endpoint"`
	Insecure  bool   `json:"insecure,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// ConfigurationReader handles reading of the storage system configuration secret
type ConfigurationReader struct{}

// GetStorageSystemConfiguration returns a storage system from the configuration file
// If no default system is supplied, the first system in the list is returned
func (c *ConfigurationReader) GetStorageSystemConfiguration(file string) ([]ArrayConnectionData, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, fmt.Errorf(fmt.Sprintf("File %s does not exist", file))
	}

	config, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("File %s errors: %v", file, err))
	}

	if string(config) == "" {
		return nil, fmt.Errorf("arrays details are not provided in vxflexos-config secret")

	}

	connectionData := make([]ArrayConnectionData, 0)
	err = json.Unmarshal(config, &connectionData)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to parse the credentials: %v", err))
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

func validateStorageSystem(system ArrayConnectionData, i int) error {
	if system.SystemID == "" {
		return fmt.Errorf(fmt.Sprintf("invalid value for system name at index %d", i))
	}
	if system.Username == "" {
		return fmt.Errorf(fmt.Sprintf("invalid value for Username at index %d", i))
	}
	if system.Password == "" {
		return fmt.Errorf(fmt.Sprintf("invalid value for Password at index %d", i))
	}
	if system.Endpoint == "" {
		return fmt.Errorf(fmt.Sprintf("invalid value for Endpoint at index %d", i))
	}
	return nil
}
