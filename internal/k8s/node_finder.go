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

package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

// NodeGetter is an interface for getting a list of storage class information
//
//go:generate mockgen -destination=mocks/node_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s NodeGetter
type NodeGetter interface {
	GetNodes() (*corev1.NodeList, error)
}

// NodeFinder is a node finder that will query the Kubernetes API for a node by its IP address
type NodeFinder struct {
	API NodeGetter
}

// GetNodes will return a kubernetes Node from an IP address
func (f *NodeFinder) GetNodes() ([]corev1.Node, error) {
	nodes, err := f.API.GetNodes()
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}
