// Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

package k8s_test

import (
	"errors"
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/dell/karavi-metrics-powerflex/internal/k8s/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_K8sStorageClassFinder(t *testing.T) {
	type checkFn func(*testing.T, []v1.StorageClass, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, storageClasses []v1.StorageClass, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []v1.StorageClass) func(t *testing.T, storageClasses []v1.StorageClass, err error) {
		return func(t *testing.T, storageClasses []v1.StorageClass, err error) {
			assert.Equal(t, expectedOutput, storageClasses)
		}
	}

	hasError := func(t *testing.T, volumes []v1.StorageClass, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (k8s.StorageClassFinder, []checkFn, *gomock.Controller){
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

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)

			finder := k8s.StorageClassFinder{API: api, DriverNames: []string{"csi-vxflexos.dellemc.com"}, StorageSystemID: "storage-system-id-1", IsDefaultStorageSystem: false}
			return finder, check(hasNoError, checkExpectedOutput(storageClasses.Items)), ctrl
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

			api.EXPECT().GetStorageClasses().Times(1).Return(storageClasses, nil)

			finder := k8s.StorageClassFinder{API: api, DriverNames: []string{"csi-vxflexos.dellemc.com", "another-csi-driver.dellemc.com"}, StorageSystemID: "storage-system-id-1", IsDefaultStorageSystem: false}
			return finder, check(hasNoError, checkExpectedOutput(storageClasses.Items)), ctrl
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

			finder := k8s.StorageClassFinder{API: api, DriverNames: []string{"csi-vxflexos.dellemc.com", "another-csi-driver.dellemc.com"}, StorageSystemID: "storage-system-id-1", IsDefaultStorageSystem: true}
			return finder, check(hasNoError, checkExpectedOutput([]v1.StorageClass{
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
						Name: "another-pool",
					},
					Provisioner: "csi-vxflexos.dellemc.com",
					Parameters: map[string]string{
						"storagepool": "mypool",
						"systemID":    "storage-system-id-1",
					},
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

			finder := k8s.StorageClassFinder{API: api, DriverNames: []string{"csi-vxflexos.dellemc.com"}, StorageSystemID: "storage-system-id-1"}
			return finder, check(hasNoError, checkExpectedOutput([]v1.StorageClass{
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
		sc := v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "vxflexos",
			},
			Provisioner: "csi-vxflexos.dellemc.com",
			Parameters: map[string]string{
				"storagepool": "mypool",
				"systemID":    "storage-system-id-1",
			},
		}

		storagePools := k8s.GetStoragePools(sc)
		expectedOutput := []string{"mypool"}
		assert.Equal(t, expectedOutput, storagePools)
	})
}
