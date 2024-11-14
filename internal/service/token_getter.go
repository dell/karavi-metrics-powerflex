// Copyright © 2021-2022 Dell Inc., or its subsidiaries. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"sync"
	"time"

	"github.com/dell/goscaleio"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// TokenManager manages and retains a valid token for a PowerFlex
type TokenManager struct {
	Config       TokenManagerConfig
	sem          chan struct{}
	mu           sync.Mutex // protects currentToken
	currentToken string
	done         chan struct{}
}

// TokenManagerConfig is the configuration for building a TokenManagerConfig
type TokenManagerConfig struct {
	PowerFlexClient      *goscaleio.Client
	TokenRefreshInterval time.Duration
	ConfigConnect        *goscaleio.ConfigConnect
	Logger               *logrus.Logger
}

// NewTokenManager returns a PowerFlexTokenGetter from the supplied Config
func NewTokenManager(c TokenManagerConfig) *TokenManager {
	return &TokenManager{
		Config: c,
		sem:    make(chan struct{}, 1),
		done:   make(chan struct{}, 1),
	}
}

// Start starts the TokenGetter to retain a valid PowerFlex token
func (tg *TokenManager) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// Update the token one time on startup, then update on timer interval after that
	tg.mu.Lock()
	tg.currentToken = ""
	tg.mu.Unlock()
	tg.updateTokenFromPowerFlex()

	timer := time.NewTimer(tg.Config.TokenRefreshInterval)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			tg.updateTokenFromPowerFlex()
			timer.Reset(tg.Config.TokenRefreshInterval)
		case <-tg.done:
			timer.Stop()
			return nil
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

// Stop stops the Token Getter
func (tg *TokenManager) Stop() {
	tg.done <- struct{}{}
}

// GetToken returns a valid token for the configured PowerFlex
func (tg *TokenManager) GetToken(ctx context.Context) (string, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("").Start(ctx, "GetToken")
	defer span.End()

	select {
	case tg.sem <- struct{}{}:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	defer func() { <-tg.sem }()
	return tg.getToken(), nil
}

func (tg *TokenManager) getToken() string {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	return tg.currentToken
}

func (tg *TokenManager) updateTokenFromPowerFlex() {
	tg.sem <- struct{}{}
	defer func() {
		<-tg.sem
	}()

	// todo: implement logout in goscaleio to log out of session before Authenticate
	if _, err := tg.Config.PowerFlexClient.Authenticate(tg.Config.ConfigConnect); err != nil {
		tg.Config.Logger.Errorf("PowerFlex Auth error: %+v", err)
	}
	tg.mu.Lock()
	tg.currentToken = tg.Config.PowerFlexClient.GetToken()
	tg.mu.Unlock()
}
