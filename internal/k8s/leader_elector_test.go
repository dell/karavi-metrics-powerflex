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

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
)

func Test_InitLeaderElection(t *testing.T) {
	type checkFn func(*testing.T, error)
	type configFn func() (*rest.Config, error)
	type clientsetFn func(config *rest.Config) (*kubernetes.Clientset, error)
	type leaderelectionFn func(lec leaderelection.LeaderElectionConfig) (*leaderelection.LeaderElector, error)
	check := func(fns ...checkFn) []checkFn { return fns }

	hasError := func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	tests := map[string]func(t *testing.T) (configFn, clientsetFn, leaderelectionFn, []checkFn){
		"error getting a valid config": func(*testing.T) (configFn, clientsetFn, leaderelectionFn, []checkFn) {
			inClusterConfig := func() (*rest.Config, error) {
				return nil, errors.New("error")
			}
			return inClusterConfig, nil, nil, check(hasError)
		},
		"error getting new clientset": func(*testing.T) (configFn, clientsetFn, leaderelectionFn, []checkFn) {
			configFn := func() (*rest.Config, error) {
				return nil, nil
			}

			clientset := func(config *rest.Config) (*kubernetes.Clientset, error) {
				return nil, errors.New("error")
			}

			return configFn, clientset, nil, check(hasError)
		},
		"error create leader": func(*testing.T) (configFn, clientsetFn, leaderelectionFn, []checkFn) {
			configFn := func() (*rest.Config, error) {
				return nil, nil
			}

			clientset := func(config *rest.Config) (*kubernetes.Clientset, error) {
				mockClientset := &kubernetes.Clientset{}
				return mockClientset, nil
			}
			leaderelection := func(lec leaderelection.LeaderElectionConfig) (*leaderelection.LeaderElector, error) {
				return nil, errors.New("error")
			}
			return configFn, clientset, leaderelection, check(hasError)
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			configFn, clientsetFn, leaderelectionFn, checkFns := tc(t)
			k8sclient := k8s.LeaderElector{}
			if configFn != nil {
				oldInClusterConfig := k8s.InClusterConfigFn
				defer func() { k8s.InClusterConfigFn = oldInClusterConfig }()
				k8s.InClusterConfigFn = configFn
			}
			if clientsetFn != nil {
				oldNewForConfigFn := k8s.NewForConfigFn
				defer func() { k8s.NewForConfigFn = oldNewForConfigFn }()
				k8s.NewForConfigFn = clientsetFn
			}
			if leaderelectionFn != nil {
				oldLeaderElection := k8s.NewLeaderElectorFn
				defer func() { k8s.NewLeaderElectorFn = oldLeaderElection }()
				k8s.NewLeaderElectorFn = leaderelectionFn
			}
			err := k8sclient.InitLeaderElection("karavi-metrics-powerflex", "karavi")
			for _, checkFn := range checkFns {
				checkFn(t, err)
			}
		})
	}
}

func Test_IsLeader(t *testing.T) {
	tt := []struct {
		Name  string
		Input bool
	}{
		{
			"Test",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			k8sclient := k8s.LeaderElector{}
			err := k8sclient.IsLeader()
			assert.Equal(t, err, tc.Input)
		})
	}
}

func Test_NewForConfigFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		k8sconfig := &rest.Config{}
		_, err := k8s.NewForConfigFn(k8sconfig)
		assert.Equal(t, err, nil)
	})
}

func Test_NewLeaderElectorFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lec := leaderelection.LeaderElectionConfig{}
		_, err := k8s.NewLeaderElectorFn(lec)
		assert.Error(t, err)
	})
}
