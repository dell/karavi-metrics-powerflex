package k8s_test

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"

	"k8s.io/client-go/kubernetes"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func Test_GetCSINodes(t *testing.T) {
	type checkFn func(*testing.T, *v1.CSINodeList, error)
	type connectFn func(*k8s.API) error
	type configFn func() (*rest.Config, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, nodes *v1.CSINodeList, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput *v1.CSINodeList) func(t *testing.T, nodes *v1.CSINodeList, err error) {
		return func(t *testing.T, nodes *v1.CSINodeList, err error) {
			assert.Equal(t, expectedOutput, nodes)
		}
	}

	hasError := func(t *testing.T, nodes *v1.CSINodeList, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (connectFn, configFn, []checkFn){
		"success": func(*testing.T) (connectFn, configFn, []checkFn) {

			nodes := &v1.CSINodeList{
				Items: []v1.CSINode{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "csi-node-1",
						},
						Spec: v1.CSINodeSpec{
							Drivers: []v1.CSINodeDriver{
								{
									Name:   "csi-vxflexos.dellemc.com",
									NodeID: "node-1",
								},
							},
						},
					},
				},
			}

			connect := func(api *k8s.API) error {
				api.Client = fake.NewSimpleClientset(nodes)
				return nil
			}
			return connect, nil, check(hasNoError, checkExpectedOutput(nodes))
		},
		"error connecting": func(*testing.T) (connectFn, configFn, []checkFn) {
			connect := func(api *k8s.API) error {
				return errors.New("error")
			}
			return connect, nil, check(hasError)
		},
		"error getting a valid config": func(*testing.T) (connectFn, configFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return nil, inClusterConfig, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			connectFn, inClusterConfig, checkFns := tc(t)
			k8sclient := k8s.API{}

			if connectFn != nil {
				oldConnectFn := k8s.ConnectFn
				defer func() { k8s.ConnectFn = oldConnectFn }()
				k8s.ConnectFn = connectFn
			}
			if inClusterConfig != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = inClusterConfig
			}
			nodes, err := k8sclient.GetCSINodes()
			for _, checkFn := range checkFns {
				checkFn(t, nodes, err)
			}
		})
	}

}

func Test_GetPersistentVolumes(t *testing.T) {
	type checkFn func(*testing.T, *corev1.PersistentVolumeList, error)
	type connectFn func(*k8s.API) error
	type configFn func() (*rest.Config, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput *corev1.PersistentVolumeList) func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		return func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
			assert.Equal(t, expectedOutput, volumes)
		}
	}

	hasError := func(t *testing.T, volumes *corev1.PersistentVolumeList, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (connectFn, configFn, []checkFn){
		"success": func(*testing.T) (connectFn, configFn, []checkFn) {

			volumes := &corev1.PersistentVolumeList{
				Items: []corev1.PersistentVolume{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "persistent-volume-name",
						},
					},
				},
			}
			connect := func(api *k8s.API) error {
				api.Client = fake.NewSimpleClientset(volumes)
				return nil
			}
			return connect, nil, check(hasNoError, checkExpectedOutput(volumes))
		},
		"error connecting": func(*testing.T) (connectFn, configFn, []checkFn) {
			connect := func(api *k8s.API) error {
				return errors.New("error")
			}
			return connect, nil, check(hasError)
		},
		"error getting a valid config": func(*testing.T) (connectFn, configFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return nil, inClusterConfig, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			connectFn, inClusterConfig, checkFns := tc(t)
			k8sclient := &k8s.API{}
			if connectFn != nil {
				oldConnectFn := k8s.ConnectFn
				defer func() { k8s.ConnectFn = oldConnectFn }()
				k8s.ConnectFn = connectFn
			}
			if inClusterConfig != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = inClusterConfig
			}
			volumes, err := k8sclient.GetPersistentVolumes()
			for _, checkFn := range checkFns {
				checkFn(t, volumes, err)
			}
		})
	}

}

func Test_GetStorageClasses(t *testing.T) {
	type checkFn func(*testing.T, *v1.StorageClassList, error)
	type connectFn func(*k8s.API) error
	type configFn func() (*rest.Config, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, volumes *v1.StorageClassList, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput *v1.StorageClassList) func(t *testing.T, volumes *v1.StorageClassList, err error) {
		return func(t *testing.T, volumes *v1.StorageClassList, err error) {
			assert.Equal(t, expectedOutput, volumes)
		}
	}

	hasError := func(t *testing.T, volumes *v1.StorageClassList, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (connectFn, configFn, []checkFn){
		"success": func(*testing.T) (connectFn, configFn, []checkFn) {
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
				},
			}

			connect := func(api *k8s.API) error {
				api.Client = fake.NewSimpleClientset(storageClasses)
				return nil
			}
			return connect, nil, check(hasNoError, checkExpectedOutput(storageClasses))
		},
		"error connecting": func(*testing.T) (connectFn, configFn, []checkFn) {
			connect := func(api *k8s.API) error {
				return errors.New("error")
			}
			return connect, nil, check(hasError)
		},
		"error getting a valid config": func(*testing.T) (connectFn, configFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return nil, inClusterConfig, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			connectFn, inClusterConfig, checkFns := tc(t)
			k8sclient := &k8s.API{}
			if connectFn != nil {
				oldConnectFn := k8s.ConnectFn
				defer func() { k8s.ConnectFn = oldConnectFn }()
				k8s.ConnectFn = connectFn
			}
			if inClusterConfig != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = inClusterConfig
			}
			storageClasses, err := k8sclient.GetStorageClasses()
			for _, checkFn := range checkFns {
				checkFn(t, storageClasses, err)
			}
		})
	}
}

func Test_GetNodes(t *testing.T) {
	type checkFn func(*testing.T, *corev1.NodeList, error)
	type connectFn func(*k8s.API) error
	type configFn func() (*rest.Config, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, nodes *corev1.NodeList, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput *corev1.NodeList) func(t *testing.T, nodes *corev1.NodeList, err error) {
		return func(t *testing.T, nodes *corev1.NodeList, err error) {
			assert.Equal(t, expectedOutput, nodes)
		}
	}

	hasError := func(t *testing.T, nodes *corev1.NodeList, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (connectFn, configFn, []checkFn){
		"success": func(*testing.T) (connectFn, configFn, []checkFn) {
			nodes := &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
					},
				},
			}

			connect := func(api *k8s.API) error {
				api.Client = fake.NewSimpleClientset(nodes)
				return nil
			}
			return connect, nil, check(hasNoError, checkExpectedOutput(nodes))
		},
		"error connecting": func(*testing.T) (connectFn, configFn, []checkFn) {
			connect := func(api *k8s.API) error {
				return errors.New("error")
			}
			return connect, nil, check(hasError)
		},
		"error getting a valid config": func(*testing.T) (connectFn, configFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return nil, inClusterConfig, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			connectFn, inClusterConfig, checkFns := tc(t)
			k8sclient := &k8s.API{}
			if connectFn != nil {
				oldConnectFn := k8s.ConnectFn
				defer func() { k8s.ConnectFn = oldConnectFn }()
				k8s.ConnectFn = connectFn
			}
			if inClusterConfig != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = inClusterConfig
			}
			nodes, err := k8sclient.GetNodes()
			for _, checkFn := range checkFns {
				checkFn(t, nodes, err)
			}
		})
	}
}

func Test_InClusterConfigFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, err := k8s.InClusterConfigFn()
		assert.Error(t, err)
	})
}

func Test_NewForConfigError(t *testing.T) {
	k8sapi := &k8s.API{}

	oldInClusterConfigFn := k8s.InClusterConfigFn
	defer func() { k8s.InClusterConfigFn = oldInClusterConfigFn }()
	k8s.InClusterConfigFn = func() (*rest.Config, error) {
		return new(rest.Config), nil
	}

	oldNewConfigFn := k8s.NewConfigFn
	defer func() { k8s.NewConfigFn = oldNewConfigFn }()
	expected := "could not create Clientset from KubeConfig"
	k8s.NewConfigFn = func(config *rest.Config) (*kubernetes.Clientset, error) {
		return nil, fmt.Errorf(expected)
	}

	_, err := k8sapi.GetStorageClasses()
	assert.True(t, err != nil)
	if err != nil {
		assert.Equal(t, expected, err.Error())
	}
}
