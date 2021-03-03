// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package k8s

import (
	v1 "k8s.io/api/storage/v1"
)

// StorageClassGetter is an interface for getting a list of storage class information
//go:generate mockgen -destination=mocks/storage_class_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s StorageClassGetter
type StorageClassGetter interface {
	GetStorageClasses() (*v1.StorageClassList, error)
}

// StorageClassFinder is a storage class finder that will query the Kubernetes API for storage classes provisioned by a matching DriverName and StorageSystemID
type StorageClassFinder struct {
	API                    StorageClassGetter
	DriverNames            []string
	StorageSystemID        string
	IsDefaultStorageSystem bool
}

// GetStorageClasses will return a list of storage classes that match the given DriverName in Kubernetes
func (f *StorageClassFinder) GetStorageClasses() ([]v1.StorageClass, error) {
	var storageClasses []v1.StorageClass

	classes, err := f.API.GetStorageClasses()
	if err != nil {
		return nil, err
	}

	for _, class := range classes.Items {
		if f.isMatch(class) {
			storageClasses = append(storageClasses, class)
		}
	}
	return storageClasses, nil
}

func (f *StorageClassFinder) isMatch(class v1.StorageClass) bool {
	if !Contains(f.DriverNames, class.Provisioner) {
		return false
	}

	systemID, systemIDExists := class.Parameters["systemID"]

	// if a storage system is marked as default, the StorageClass is a match if either the 'systemID' key does not exist or if it matches the storage system ID
	if f.IsDefaultStorageSystem {
		return !systemIDExists || systemID == f.StorageSystemID
	}

	// if a storage system is not marked as default, the StorageClass is a match only if the 'systemID' matches the storage system ID
	return systemID == f.StorageSystemID
}

// GetStoragePools will return a list of storage pool names from a given Kubernetes storage class
func GetStoragePools(storageClass v1.StorageClass) []string {
	return []string{storageClass.Parameters["storagepool"]}
}
