package k8s

// Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// API holds data used to access the K8S API
type API struct {
	Client kubernetes.Interface
	Lock   sync.Mutex
}

// GetCSINodes will return a list of CSI nodes in the kubernetes cluster
func (api *API) GetCSINodes() (*v1.CSINodeList, error) {
	api.Lock.Lock()
	defer api.Lock.Unlock()
	if api.Client == nil {
		err := ConnectFn(api)
		if err != nil {
			return nil, err
		}
	}
	return api.Client.StorageV1beta1().CSINodes().List(metav1.ListOptions{})
}

// GetPersistentVolumes will return a list of persistent volumes in the kubernetes cluster
func (api *API) GetPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	api.Lock.Lock()
	defer api.Lock.Unlock()
	if api.Client == nil {
		err := ConnectFn(api)
		if err != nil {
			return nil, err
		}
	}
	return api.Client.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}

// GetStorageClasses will return a list of storage classes in the kubernetes clusteer
func (api *API) GetStorageClasses() (*v1.StorageClassList, error) {
	api.Lock.Lock()
	defer api.Lock.Unlock()
	if api.Client == nil {
		err := ConnectFn(api)
		if err != nil {
			return nil, err
		}
	}
	return api.Client.StorageV1beta1().StorageClasses().List(metav1.ListOptions{})
}

// GetNodes will return the list of nodes in the kubernetes cluster
func (api *API) GetNodes() (*corev1.NodeList, error) {
	api.Lock.Lock()
	defer api.Lock.Unlock()
	if api.Client == nil {
		err := ConnectFn(api)
		if err != nil {
			return nil, err
		}
	}

	return api.Client.CoreV1().Nodes().List(metav1.ListOptions{})
}

// ConnectFn will connect the client to the k8s API
var ConnectFn = func(api *API) error {
	config, err := getConfig()
	if err != nil {
		return err
	}
	api.Client, err = NewConfigFn(config)
	if err != nil {
		return err
	}
	return nil
}

// InClusterConfigFn will return a valid configuration if we are running in a Pod on a kubernetes cluster
var InClusterConfigFn = func() (*rest.Config, error) {
	return rest.InClusterConfig()
}

// NewConfigFn will return a valid kubernetes.Clientset
var NewConfigFn = func(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

func getConfig() (*rest.Config, error) {
	config, err := InClusterConfigFn()
	if err != nil {
		return nil, err
	}
	return config, nil
}
