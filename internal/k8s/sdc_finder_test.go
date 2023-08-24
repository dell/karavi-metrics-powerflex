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

package k8s_test

import (
	"errors"
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s/mocks"

	v1 "k8s.io/api/storage/v1"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_K8sSDCFinder(t *testing.T) {
	type checkFn func(*testing.T, []string, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, sdcGuids []string, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []string) func(t *testing.T, sdcGuids []string, err error) {
		return func(t *testing.T, sdcGuids []string, err error) {
			assert.Equal(t, expectedOutput, sdcGuids)
		}
	}

	hasError := func(t *testing.T, sdcGuids []string, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (k8s.SDCFinder, []checkFn, *gomock.Controller){
		"success": func(*testing.T) (k8s.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockKubernetesAPI(ctrl)

			nodes := &v1.CSINodeList{
				Items: []v1.CSINode{
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "csi-vxflexos.dellemc.com",
									NodeID:       "node-1",
									TopologyKeys: []string{"csi-vxflexos.dellemc.com/storage-system-id-1"},
								},
							},
						},
					},
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:   "other-driver-name",
									NodeID: "node-2",
								},
							},
						},
					},
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "csi-vxflexos.dellemc.com",
									NodeID:       "node-3",
									TopologyKeys: []string{"csi-vxflexos.dellemc.com/storage-system-id-1"},
								},
							},
						},
					},
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "csi-vxflexos.dellemc.com",
									NodeID:       "node-4",
									TopologyKeys: []string{"csi-vxflexos.dellemc.com/storage-system-id-2"},
								},
							},
						},
					},
				},
			}
			api.EXPECT().GetCSINodes().Times(1).Return(nodes, nil)

			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com"}}

			finder := k8s.SDCFinder{API: api, StorageSystemID: ids}
			return finder, check(hasNoError, checkExpectedOutput([]string{"node-1", "node-3"})), ctrl
		},
		"success with multiple driver names": func(*testing.T) (k8s.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockKubernetesAPI(ctrl)

			nodes := &v1.CSINodeList{
				Items: []v1.CSINode{
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "csi-vxflexos.dellemc.com",
									NodeID:       "node-1",
									TopologyKeys: []string{"csi-vxflexos.dellemc.com/storage-system-id-1"},
								},
							},
						},
					},
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "other-driver-name",
									NodeID:       "node-2",
									TopologyKeys: []string{"other-driver-name/storage-system-id-1"},
								},
							},
						},
					},
					{
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:         "csi-vxflexos.dellemc.com",
									NodeID:       "node-3",
									TopologyKeys: []string{"csi-vxflexos.dellemc.com/storage-system-id-1"},
								},
							},
						},
					},
				},
			}
			api.EXPECT().GetCSINodes().Times(1).Return(nodes, nil)

			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com", "other-driver-name"}}

			finder := k8s.SDCFinder{API: api, StorageSystemID: ids}

			return finder, check(hasNoError, checkExpectedOutput([]string{"node-1", "node-2", "node-3"})), ctrl
		},
		"error calling k8s": func(*testing.T) (k8s.SDCFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockKubernetesAPI(ctrl)
			api.EXPECT().GetCSINodes().Times(1).Return(nil, errors.New("error"))
			finder := k8s.SDCFinder{API: api}
			return finder, check(hasError), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			finder, checkFns, ctrl := tc(t)
			sdcGuids, err := finder.GetSDCGuids()
			for _, checkFn := range checkFns {
				checkFn(t, sdcGuids, err)
			}
			ctrl.Finish()
		})
	}
}
