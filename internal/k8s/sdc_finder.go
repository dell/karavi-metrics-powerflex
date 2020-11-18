package k8s

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	v1 "k8s.io/api/storage/v1beta1"
)

// KubernetesAPI is an interface for accessing the Kubernetes API
//go:generate mockgen -destination=mocks/kubernetes_api_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s KubernetesAPI
type KubernetesAPI interface {
	GetCSINodes() (*v1.CSINodeList, error)
}

// SDCFinder is an SDC finder that will query the Kubernetes API for CSI-Nodes that have a matching DriverName
type SDCFinder struct {
	API         KubernetesAPI
	DriverNames []string
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
			if Contains(f.DriverNames, driver.Name) {
				sdcGUIDS = append(sdcGUIDS, driver.NodeID)
			}
		}
	}
	return sdcGUIDS, nil
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
