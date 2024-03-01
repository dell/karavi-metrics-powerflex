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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_K8sNodeFinder(t *testing.T) {
	type checkFn func(*testing.T, []corev1.Node, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasNoError := func(t *testing.T, _ []corev1.Node, err error) {
		if err != nil {
			t.Fatalf("expected no error")
		}
	}

	checkExpectedOutput := func(expectedOutput []corev1.Node) func(t *testing.T, nodes []corev1.Node, err error) {
		return func(t *testing.T, nodes []corev1.Node, _ error) {
			assert.Equal(t, expectedOutput, nodes)
		}
	}

	hasError := func(t *testing.T, _ []corev1.Node, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (k8s.NodeFinder, []checkFn, *gomock.Controller){
		"success finding nodes": func(*testing.T) (k8s.NodeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockNodeGetter(ctrl)

			nodes := &corev1.NodeList{
				Items: []corev1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node1",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.4",
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "node2",
						},
						Status: corev1.NodeStatus{
							Addresses: []corev1.NodeAddress{
								{
									Address: "1.2.3.5",
								},
							},
						},
					},
				},
			}

			api.EXPECT().GetNodes().Times(1).Return(nodes, nil)

			finder := k8s.NodeFinder{API: api}
			return finder, check(hasNoError, checkExpectedOutput(nodes.Items)), ctrl
		},
		"error calling k8s": func(*testing.T) (k8s.NodeFinder, []checkFn, *gomock.Controller) {
			ctrl := gomock.NewController(t)
			api := mocks.NewMockNodeGetter(ctrl)
			api.EXPECT().GetNodes().Times(1).Return(nil, errors.New("error"))
			finder := k8s.NodeFinder{API: api}
			return finder, check(hasError), ctrl
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			finder, checkFns, ctrl := tc(t)
			nodes, err := finder.GetNodes()
			for _, checkFn := range checkFns {
				checkFn(t, nodes, err)
			}
			ctrl.Finish()
		})
	}
}
