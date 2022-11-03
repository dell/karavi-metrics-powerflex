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
	"context"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// LeaderElectorGetter is an interface for initialize and check elected leader
//
//go:generate mockgen -destination=mocks/leader_elector_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/k8s LeaderElectorGetter
type LeaderElectorGetter interface {
	InitLeaderElection(string, string) error
	IsLeader() bool
}

// LeaderElector holds LeaderElector struct for the client
type LeaderElector struct {
	API     LeaderElectorGetter
	Elector *leaderelection.LeaderElector
}

// InitLeaderElection will run algorithm for leader election, call during service initialzation process
func (elect *LeaderElector) InitLeaderElection(endpoint string, namespace string) error {

	k8sconfig, err := InClusterConfigFn()
	if err != nil {
		return err
	}

	k8sclient, err := NewForConfigFn(k8sconfig)
	if err != nil {
		return err
	}

	leaderConfig := leaderelection.LeaderElectionConfig{
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(context.Context) {},
			OnStoppedLeading: func() {},
			OnNewLeader:      func(identity string) {},
		},
		Lock: &resourcelock.EndpointsLock{
			EndpointsMeta: metav1.ObjectMeta{
				Name:      endpoint,
				Namespace: namespace,
			},
			Client: k8sclient.CoreV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: os.Getenv("HOSTNAME"),
			},
		},
	}

	elect.Elector, err = NewLeaderElectorFn(leaderConfig)
	if err != nil {
		return err
	}

	elect.Elector.Run(context.Background())
	return nil
}

// IsLeader return true if the given client is leader at the moment
func (elect *LeaderElector) IsLeader() bool {
	if elect.Elector == nil {
		return false
	}
	return elect.Elector.IsLeader()
}

// NewForConfigFn creates a new Clientset for the given config. If config's RateLimiter is not set and QPS and Burst are acceptable, NewForConfigFn will generate a rate-limiter in configShallowCopy
var NewForConfigFn = func(k8sconfig *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(k8sconfig)
}

// NewLeaderElectorFn creates a LeaderElector from a LeaderElectionConfig
var NewLeaderElectorFn = func(lec leaderelection.LeaderElectionConfig) (*leaderelection.LeaderElector, error) {
	return leaderelection.NewLeaderElector(lec)
}
