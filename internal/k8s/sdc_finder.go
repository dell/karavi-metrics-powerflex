// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package k8s

import (
	"strings"

	v1 "k8s.io/api/storage/v1"
)

// KubernetesAPI is an interface for accessing the Kubernetes API
//
//go:generate mockgen -destination=mocks/kubernetes_api_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s KubernetesAPI
type KubernetesAPI interface {
	GetCSINodes() (*v1.CSINodeList, error)
}

// SDCFinder is an SDC finder that will query the Kubernetes API for CSI-Nodes that have a matching DriverName and Storage System ID
type SDCFinder struct {
	API             KubernetesAPI
	StorageSystemID []StorageSystemID
}

// GetSDCGuids will return a list of SDC GUIDs that match the given DriverName in Kubernetes
func (f *SDCFinder) GetSDCGuids() ([]string, error) {
	var sdcGUIDS []string

	nodes, err := f.API.GetCSINodes()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		for _, driver := range node.Spec.Drivers {
			if f.isMatch(driver) {
				sdcGUIDS = append(sdcGUIDS, driver.NodeID)
			}
		}
	}
	return sdcGUIDS, nil
}

func (f *SDCFinder) isMatch(driver v1.CSINodeDriver) bool {
	for _, topologyKey := range driver.TopologyKeys {
		split := strings.Split(topologyKey, "/")
		if len(split) == 2 {
			for _, storage := range f.StorageSystemID {
				if split[1] == storage.ID && Contains(storage.DriverNames, split[0]) {
					return true
				}
			}

		}
	}
	return false
}

// Contains will return true if the slice contains the given value
func Contains(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}
