/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

import types "github.com/dell/goscaleio/types/v1"

// MappedSDC is the summerized details of the SDCs volume is mapped to
type MappedSDC struct {
	SdcID string `json:"sdcId"`
	SdcIP string `json:"sdcIp"`
}

// VolumeMeta is the details of a volume in an SDC
type VolumeMeta struct {
	ID                        string
	Name                      string
	PersistentVolumeName      string
	PersistentVolumeClaimName string
	Namespace                 string
	StorageSystemID           string
	MappedSDCs                []MappedSDC
}

// VolumeMetaMetrics is the details of a volume in an SDC along with the metrics
type VolumeMetaMetrics struct {
	ID                        string
	Name                      string
	PersistentVolumeName      string
	PersistentVolumeClaimName string
	Namespace                 string
	StorageSystemID           string
	MappedSDCs                []MappedSDC
	ReadLatencyBwc            types.BWC
	ReadBwc                   types.BWC
	TrimBwc                   types.BWC
	TrimLatencyBwc            types.BWC
	WriteBwc                  types.BWC
	WriteLatencyBwc           types.BWC
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

type TopologyMeta struct {
	Namespace               string
	PersistentVolumeClaim   string
	PersistentVolumeStatus  string
	VolumeClaimName         string
	PersistentVolume        string
	StorageClass            string
	Driver                  string
	ProvisionedSize         string
	StorageSystemVolumeName string
	StoragePoolName         string
	StorageSystem           string
	Protocol                string
	CreatedTime             string
}
