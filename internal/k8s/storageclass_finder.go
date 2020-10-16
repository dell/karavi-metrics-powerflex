package k8s

import (
	v1 "k8s.io/api/storage/v1beta1"
)

// StorageClassGetter is an interface for getting a list of storage class information
//go:generate mockgen -destination=mocks/storage_class_getter_mocks.go -package=mocks github.com/dell/karavi-powerflex-metrics/internal/k8s StorageClassGetter
type StorageClassGetter interface {
	GetStorageClasses() (*v1.StorageClassList, error)
}

// StorageClassFinder is a storage class finder that will query the Kubernetes API for storage classes provisioned by a matching DriverName
type StorageClassFinder struct {
	API         StorageClassGetter
	DriverNames []string
}

// GetStorageClasses will return a list of storage classes that match the given DriverName in Kubernetes
func (f *StorageClassFinder) GetStorageClasses() ([]v1.StorageClass, error) {
	var storageClasses []v1.StorageClass

	classes, err := f.API.GetStorageClasses()
	if err != nil {
		return nil, err
	}

	for _, class := range classes.Items {
		if Contains(f.DriverNames, class.Provisioner) {
			storageClasses = append(storageClasses, class)
		}
	}
	return storageClasses, nil
}

// GetStoragePools will return a list of storage pool names from a given Kubernetes storage class
func GetStoragePools(storageClass v1.StorageClass) []string {
	return []string{storageClass.Parameters["storagepool"]}
}
