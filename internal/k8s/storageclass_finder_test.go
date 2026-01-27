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

	"github.com/dell/karavi-metrics-powerflex/internal/domain"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_K8sStorageClassFinder(t *testing.T) {
	type checkFn func(*testing.T, []k8s.StorageClass, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, _ []k8s.StorageClass, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []k8s.StorageClass) func(t *testing.T, storageClasses []k8s.StorageClass, err error) {
		return func(t *testing.T, storageClasses []k8s.StorageClass, _ error) {
			assert.Equal(t, expectedOutput, storageClasses)
		}
	}

	hasError := func(t *testing.T, _ []k8s.StorageClass, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller){
		"success not selecting storageclass that is not in config": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "notExisting",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
				},
			}

			expected := []k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)
			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com"}, IsDefault: false}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}
			return finder, check(hasNoError, checkExpectedOutput(expected)), ctrl
		},
		"success selecting the matching driver name with storage classes": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
				},
			}

			expected := []k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)
			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com"}, IsDefault: false}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}
			return finder, check(hasNoError, checkExpectedOutput(expected)), ctrl
		},
		"success selecting storage classes matching multiple driver names": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-another",
						},
						Provisioner: "another-csi-driver.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs-another",
						},
						Provisioner: "another-csi-driver.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
				},
			}

			expected := []k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-another",
						},
						Provisioner: "another-csi-driver.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs-another",
						},
						Provisioner: "another-csi-driver.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)
			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com", "another-csi-driver.dellemc.com"}, IsDefault: false}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}

			return finder, check(hasNoError, checkExpectedOutput(expected)), ctrl
		},
		"success matching storage classes without systemID based on a default system being used": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-2",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "another-pool",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)

			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com", "another-csi-driver.dellemc.com"}, IsDefault: true}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}

			return finder, check(hasNoError, checkExpectedOutput([]k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "another-pool",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
			},
			)), ctrl
		},
		"success selecting storage classes matching one of two driver names": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-another",
						},
						Provisioner: "another-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs-another",
						},
						Provisioner: "another-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)

			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com"}}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}
			return finder, check(hasNoError, checkExpectedOutput([]k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos-xfs",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						Parameters: map[string]string{
							"storagepool": "mypool",
							"systemID":    "storage-system-id-1",
						},
					},
					SystemID: "storage-system-id-1",
				},
			},
			)), ctrl
		},
		"success matching storage classes without systemID based on availability zone": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)

			storageClasses := &v1.StorageClassList{
				Items: []v1.StorageClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						AllowedTopologies: []corev1.TopologySelectorTerm{
							{
								MatchLabelExpressions: []corev1.TopologySelectorLabelRequirement{
									{
										Key:    "topology.kubernetes.io/zone",
										Values: []string{"zone1"},
									},
								},
							},
						},
					},
				},
			}

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)

			ids := make([]k8s.StorageSystemID, 1)
			ids[0] = k8s.StorageSystemID{ID: "storage-system-id-1", DriverNames: []string{"csi-vxflexos.dellemc.com"}, AvailabilityZone: &domain.AvailabilityZone{
				Name:     "zone1",
				LabelKey: "topology.kubernetes.io/zone",
				ProtectionDomains: []domain.ProtectionDomain{
					{
						Name: "zone1",
					},
				},
			}}

			finder := k8s.StorageClassFinder{API: api, StorageSystemID: ids}

			return finder, check(hasNoError, checkExpectedOutput([]k8s.StorageClass{
				{
					StorageClass: v1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "vxflexos",
						},
						Provisioner: "csi-vxflexos.dellemc.com",
						AllowedTopologies: []corev1.TopologySelectorTerm{
							{
								MatchLabelExpressions: []corev1.TopologySelectorLabelRequirement{
									{
										Key:    "topology.kubernetes.io/zone",
										Values: []string{"zone1"},
									},
								},
							},
						},
					},
					SystemID: "storage-system-id-1",
				},
			},
			)), ctrl
		},
		"error calling k8s": func(*testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockStorageClassGetter(ctrl)
			api.EXPECT().GetStorageClasses().Times(1).Return(nil, errors.New("error"))
			finder := k8s.StorageClassFinder{API: api}
			return finder, check(hasError), ctrl
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			finder, checkFns, ctrl := tc(t)
			storageClasses, err := finder.GetStorageClasses()
			for _, checkFn := range checkFns {
				checkFn(t, storageClasses, err)
			}
			ctrl.Finish()
		})
	}
}

func Test_GetStoragePools(t *testing.T) {
	t.Run("success, single storage pool retrieved from storage class", func(t *testing.T) {
		sc := k8s.StorageClass{
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "vxflexos",
				},
				Provisioner: "csi-vxflexos.dellemc.com",
				Parameters: map[string]string{
					"storagepool": "mypool",
					"systemID":    "storage-system-id-1",
				},
			},
			SystemID: "storage-system-id-1",
		}

		scf := k8s.StorageClassFinder{}

		storagePools := scf.GetStoragePools(sc)
		expectedOutput := []string{"mypool"}
		assert.Equal(t, expectedOutput, storagePools)
	})

	t.Run("it gets the storage pools from the availability zone", func(t *testing.T) {
		sc := k8s.StorageClass{
			StorageClass: v1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "vxflexos",
				},
				Provisioner: "csi-vxflexos.dellemc.com",
				Parameters:  map[string]string{},
			},
			SystemID: "storage-system-id-1",
		}

		scf := k8s.StorageClassFinder{
			StorageSystemID: []k8s.StorageSystemID{
				{
					ID: "storage-system-id-1",
					AvailabilityZone: &domain.AvailabilityZone{
						ProtectionDomains: []domain.ProtectionDomain{
							{
								Pools: []domain.PoolName{
									"mypool",
								},
							},
						},
					},
				},
			},
		}

		storagePools := scf.GetStoragePools(sc)
		expectedOutput := []string{"mypool"}
		assert.Equal(t, expectedOutput, storagePools)
	})
}
