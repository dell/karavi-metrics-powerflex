package k8s

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

// VolumeGetter is an interface for getting a list of persistent volume information
//go:generate mockgen -destination=mocks/volume_getter_mocks.go -package=mocks github.com/dell/karavi-powerflex-metrics/internal/k8s VolumeGetter
type VolumeGetter interface {
	GetPersistentVolumes() (*corev1.PersistentVolumeList, error)
}

// VolumeFinder is a volume finder that will query the Kubernetes API for Persistent Volumes created by a matching DriverName
type VolumeFinder struct {
	API         VolumeGetter
	DriverNames []string
}

// VolumeInfo contains information about mapping a Persistent Volume to the volume created on a storage system
type VolumeInfo struct {
	Namespace               string `json:"namespace"`
	PersistentVolumeClaim   string `json:"persistent_volume_claim"`
	PersistentVolumeStatus  string `json:"volume_status"`
	VolumeClaimName         string `json:"volume_claim_name"`
	PersistentVolume        string `json:"persistent_volume"`
	StorageClass            string `json:"storage_class"`
	Driver                  string `json:"driver"`
	ProvisionedSize         string `json:"provisioned_size"`
	StorageSystemVolumeName string `json:"storage_system_volume_name"`
	StoragePoolName         string `json:"storage_pool_name"`
	CreatedTime             string `json:"created_time"`
}

// GetPersistentVolumes will return a list of persistent volume information
func (f VolumeFinder) GetPersistentVolumes() ([]VolumeInfo, error) {
	volumeInfo := make([]VolumeInfo, 0)

	volumes, err := f.API.GetPersistentVolumes()
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes.Items {

		if Contains(f.DriverNames, volume.Spec.CSI.Driver) {
			capacity := volume.Spec.Capacity[v1.ResourceStorage]
			claim := volume.Spec.ClaimRef
			status := volume.Status

			info := VolumeInfo{
				Namespace:               claim.Namespace,
				PersistentVolumeClaim:   string(claim.UID),
				VolumeClaimName:         claim.Name,
				PersistentVolumeStatus:  string(status.Phase),
				PersistentVolume:        volume.Name,
				StorageClass:            volume.Spec.StorageClassName,
				Driver:                  volume.Spec.CSI.Driver,
				ProvisionedSize:         capacity.String(),
				StorageSystemVolumeName: volume.Spec.CSI.VolumeAttributes["Name"],
				StoragePoolName:         volume.Spec.CSI.VolumeAttributes["StoragePoolName"],
				CreatedTime:             volume.CreationTimestamp.String(),
			}
			volumeInfo = append(volumeInfo, info)
		}
	}
	return volumeInfo, nil
}
