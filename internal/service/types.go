package service

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

// MappedSDC is the summerized details of the SDCs volume is mapped to
type MappedSDC struct {
	SdcID string `json:"sdcId"`
	SdcIP string `json:"sdcIp"`
}

// VolumeMeta is the details of a volume in an SDC
type VolumeMeta struct {
	ID                   string
	Name                 string
	PersistentVolumeName string
	StorageSystemID      string
	MappedSDCs           []MappedSDC
}

// SDCMeta is meta data for a specific SDC
type SDCMeta struct {
	ID      string
	Name    string
	IP      string
	SdcGUID string
}

// StorageClassInfo is meta data about a storage class and contains the associated PowerFlex storage pool names
type StorageClassInfo struct {
	ID              string
	Name            string
	Driver          string
	StorageSystemID string
	StoragePools    []string
}

// StorageClassMeta is the same as StorageClassInfo except it contains a map of PowerFlex storage pool IDs to goscaleio storage pool structs
type StorageClassMeta struct {
	ID              string
	Name            string
	Driver          string
	StorageSystemID string
	StoragePools    map[string]StoragePoolStatisticsGetter
}
