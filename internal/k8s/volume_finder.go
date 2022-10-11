// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package k8s

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

// VolumeGetter is an interface for getting a list of persistent volume information
//
//go:generate mockgen -destination=mocks/volume_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s VolumeGetter
type VolumeGetter interface {
	GetPersistentVolumes() (*corev1.PersistentVolumeList, error)
}

// VolumeFinder is a volume finder that will query the Kubernetes API for Persistent Volumes created by a matching DriverName and StorageSystemID
type VolumeFinder struct {
	API             VolumeGetter
	StorageSystemID []StorageSystemID
	Logger          *logrus.Logger
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
	StorageSystemID         string `json:"storage_system_id"`
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
		if f.isMatch(volume) {
			capacity := volume.Spec.Capacity[v1.ResourceStorage]
			claim := volume.Spec.ClaimRef
			status := volume.Status
			storageystemid, err := f.getStorageID(volume)
			if err != nil {
				f.Logger.WithField("volume name", volume.Name).Warn("no storage system id found")
				continue
			}

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
				StorageSystemID:         storageystemid,
				CreatedTime:             volume.CreationTimestamp.String(),
			}
			volumeInfo = append(volumeInfo, info)
		}
	}
	return volumeInfo, nil
}

func (f *VolumeFinder) isMatch(volume v1.PersistentVolume) bool {
	if volume.Spec.CSI == nil {
		return false
	}
	// volumeHandle is storageSystemID-volumeID
	volstorageid, err := f.getStorageID(volume)
	if err != nil {
		f.Logger.WithField("volume name", volume.Name).Warn("no storage system id found")
		return false
	}
	for _, storageSystemID := range f.StorageSystemID {
		if volstorageid == storageSystemID.ID && Contains(storageSystemID.DriverNames, volume.Spec.CSI.Driver) {
			return true
		}
	}
	return false
}

func (f *VolumeFinder) getStorageID(volume v1.PersistentVolume) (string, error) {
	if volume.Spec.CSI == nil {
		return "", errors.New("storage system id not found")
	}
	// volumeHandle is storageSystemID-volumeID
	split := strings.Split(volume.Spec.CSI.VolumeHandle, "-")
	if len(split) == 2 {
		return split[0], nil
	}
	return "", errors.New("storage system id not found")
}
