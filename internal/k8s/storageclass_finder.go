/*
 Copyright (c) 2020-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package k8s

import (
	"slices"

	"github.com/dell/karavi-metrics-powerflex/internal/domain"
	v1 "k8s.io/api/storage/v1"
)

// StorageClassGetter is an interface for getting a list of storage class information
//
//go:generate mockgen -destination=mocks/storage_class_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s StorageClassGetter
type StorageClassGetter interface {
	GetStorageClasses() (*v1.StorageClassList, error)
}

// StorageSystemID contains ID, whether is default and associated drivernames
type StorageSystemID struct {
	ID               string
	IsDefault        bool
	DriverNames      []string
	AvailabilityZone *domain.AvailabilityZone
}

// StorageClassFinder is a storage class finder that will query the Kubernetes API for storage classes provisioned by a matching DriverName and StorageSystemID
type StorageClassFinder struct {
	API             StorageClassGetter
	StorageSystemID []StorageSystemID
}

// StorageClass wraps a kubernetes StorageClass to include a SystemID
type StorageClass struct {
	v1.StorageClass
	SystemID string
}

// GetStorageClasses will return a list of storage classes that match the given DriverName in Kubernetes
func (f *StorageClassFinder) GetStorageClasses() ([]StorageClass, error) {
	var storageClasses []StorageClass

	classes, err := f.API.GetStorageClasses()
	if err != nil {
		return nil, err
	}

	for _, class := range classes.Items {
		if sc := f.isMatch(class); sc != nil {
			storageClasses = append(storageClasses, *sc)
		}
	}
	return storageClasses, nil
}

func (f *StorageClassFinder) isMatch(class v1.StorageClass) *StorageClass {
	for _, storage := range f.StorageSystemID {
		if !Contains(storage.DriverNames, class.Provisioner) {
			continue
		}

		systemID, ok := class.Parameters["systemID"]
		// if the systemID field is not in the StorageClass, this could be a multi-az configuration
		// the StorageClass is a match if a zone in the storage class is associated with the StorageSystemID
		if !ok && storage.AvailabilityZone != nil {
			zone := string(storage.AvailabilityZone.Name)
			labelKey := string(storage.AvailabilityZone.LabelKey)

			for _, allowedTopology := range class.AllowedTopologies {
				for _, matchedLabelExpression := range allowedTopology.MatchLabelExpressions {
					if matchedLabelExpression.Key == labelKey {
						if slices.Contains(matchedLabelExpression.Values, zone) {
							return &StorageClass{class, storage.ID}
						}
					}
				}
			}

		}

		if systemID == storage.ID {
			return &StorageClass{class, storage.ID}
		}
	}

	for _, storage := range f.StorageSystemID {
		if !Contains(storage.DriverNames, class.Provisioner) {
			continue
		}

		systemID, systemIDExists := class.Parameters["systemID"]
		// if a storage system is marked as default, the StorageClass is a match if either the 'systemID' key does not exist or if it matches the storage system ID
		if storage.IsDefault && (!systemIDExists || systemID == storage.ID) {
			return &StorageClass{class, storage.ID}
		}
	}

	return false
}

// GetStoragePools will return a list of storage pool names from a given Kubernetes storage class
func (f *StorageClassFinder) GetStoragePools(storageClass StorageClass) []string {
	pool, ok := storageClass.Parameters["storagepool"]
	if ok {
		return []string{pool}
	}

	// if the storagepool is not in the StorageClass, this is a multi-az configuration
	// the pools must be gathered from the availablity zone
	pools := []string{}
	for _, storage := range f.StorageSystemID {
		if storage.ID == storageClass.SystemID {
			for _, protectionDomain := range storage.AvailabilityZone.ProtectionDomains {
				for _, pool := range protectionDomain.Pools {
					pools = append(pools, string(pool))
				}
			}
		}
	}
	return pools
}
