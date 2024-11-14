// Copyright © 2024 Dell Inc., or its subsidiaries. All Rights Reserved.
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

package service_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/service"

	"github.com/dell/goscaleio"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func TestLogin_GetToken(t *testing.T) {
	t.Run("success getting a token", func(t *testing.T) {
		tokens := make(map[string]interface{})
		credFile, err := os.ReadFile("testdata/tokens.yaml")
		if err != nil {
			t.Errorf("unable to read token: %v", err)
		}
		err = yaml.Unmarshal(credFile, &tokens)
		if err != nil {
			t.Errorf("unable to unmarshal token: %v", err)
		}
		firstToken := tokens["firstToken"].(string)
		// Arrange

		// Ready channel to know when tokengetter is ready
		ready := make(chan struct{})

		// Setup httptest server to represent a PowerFlex
		powerFlexSvr := newPowerFlexTestServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/api/version":
				w.Write([]byte("3.5"))
			case "/api/login":
				w.Write([]byte(firstToken))
				ready <- struct{}{}
			default:
				panic(fmt.Sprintf("path %s not supported", r.URL.String()))
			}
		})
		defer powerFlexSvr.Close()

		// Create a new TokenGetter pointing to the httptest server PowerFlex
		// TokenRefreshInterval shouldn't be relevant in this test case
		config := service.TokenManagerConfig{
			PowerFlexClient:      newPowerFlexClient(t, powerFlexSvr.URL),
			TokenRefreshInterval: time.Minute,
			Logger:               logrus.WithTime(time.Now()).Logger,
			ConfigConnect: &goscaleio.ConfigConnect{
				Endpoint: powerFlexSvr.URL,
				Version:  "",
				Username: "Test",
				Password: "Test",
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lh := service.NewTokenManager(config)
		go lh.Start(ctx)
		<-ready

		// Act

		// Get a token
		token, err := lh.GetToken(context.Background())

		// Assert

		// Assert that the token we got is the expected token from the httptest server PowerFlex
		if token != firstToken {
			t.Errorf("expected token %s, got %s", firstToken, token)
		}

		// Assert that err is nil
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("success getting a token during refresh", func(t *testing.T) {
		tokens := make(map[string]interface{})
		credFile, err := os.ReadFile("testdata/tokens.yaml")
		if err != nil {
			t.Errorf("unable to read token: %v", err)
		}
		err = yaml.Unmarshal(credFile, &tokens)
		if err != nil {
			t.Errorf("unable to unmarshal token: %v", err)
		}
		firstToken := tokens["firstToken"].(string)
		secondToken := tokens["secondToken"].(string)
		// Arrange

		// Ready channel to know when tokengetter is ready
		ready := make(chan struct{})

		// Variable to keep track of the /api/login call count so we can return different things in the following httptest server
		powerFlexCallCount := 0

		// Setup httptest server to represent a PowerFlex
		powerFlexSvr := newPowerFlexTestServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/api/version":
				w.Write([]byte("3.5"))
			case "/api/login":
				switch powerFlexCallCount {
				case 0:
					w.Write([]byte(firstToken))
					ready <- struct{}{}
					powerFlexCallCount++
				case 1:
					// Sleep to simulate this call taking longer
					time.Sleep(2 * time.Second)
					w.Write([]byte(secondToken))
				default:
					panic("unexpected call to httptest server")
				}
			default:
				panic(fmt.Sprintf("path %s not supported", r.URL.String()))
			}
		})
		defer powerFlexSvr.Close()

		// Create a new TokenGetter pointing to the httptest server PowerFlex
		config := service.TokenManagerConfig{
			PowerFlexClient:      newPowerFlexClient(t, powerFlexSvr.URL),
			TokenRefreshInterval: time.Second,
			Logger:               logrus.WithTime(time.Now()).Logger,
			ConfigConnect: &goscaleio.ConfigConnect{
				Endpoint: powerFlexSvr.URL,
				Version:  "",
				Username: "Test",
				Password: "Test",
			},
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lh := service.NewTokenManager(config)
		go lh.Start(ctx)
		<-ready

		// Act

		// Wait for refresh interval to start
		<-time.After(2 * time.Second)

		// Get a token while TokenGetter is refreshing
		token, err := lh.GetToken(context.Background())

		// Assert

		// Assert that the token we got is the expected token from the httptest server PowerFlex
		if token != secondToken {
			t.Errorf("expected token %s, got %s", secondToken, token)
		}

		// Assert that err is nil
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("timeout getting a token during refresh", func(t *testing.T) {
		tokens := make(map[string]interface{})
		credFile, err := os.ReadFile("testdata/tokens.yaml")
		if err != nil {
			t.Errorf("unable to read token: %v", err)
		}
		err = yaml.Unmarshal(credFile, &tokens)
		if err != nil {
			t.Errorf("unable to unmarshal token: %v", err)
		}
		firstToken := tokens["firstToken"].(string)
		// Arrange

		// Ready channel to know when tokengetter is ready
		ready := make(chan struct{})

		// Variable to keep track of the /api/login call count so we can return different things in the following httptest server
		powerFlexCallCount := 0

		// Setup httptest server to represent a PowerFlex
		powerFlexSvr := newPowerFlexTestServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case "/api/version":
				w.Write([]byte("3.5"))
			case "/api/login":
				switch powerFlexCallCount {
				case 0:
					w.Write([]byte(firstToken))
					ready <- struct{}{}
					powerFlexCallCount++
				case 1:
					// Sleep to simulate this call taking longer
					time.Sleep(5 * time.Second)
				default:
					panic("unexpected call to httptest server")
				}
			default:
				panic(fmt.Sprintf("path %s not supported", r.URL.String()))
			}
		})
		defer powerFlexSvr.Close()

		// Create a new TokenGetter pointing to the httptest server PowerFlex
		config := service.TokenManagerConfig{
			PowerFlexClient:      newPowerFlexClient(t, powerFlexSvr.URL),
			TokenRefreshInterval: time.Second,
			Logger:               logrus.WithTime(time.Now()).Logger,
			ConfigConnect: &goscaleio.ConfigConnect{
				Endpoint: powerFlexSvr.URL,
				Version:  "",
				Username: "Test",
				Password: "Test",
			},
		}
		ctx, cancelTokenGetter := context.WithCancel(context.Background())
		defer cancelTokenGetter()
		lh := service.NewTokenManager(config)
		go lh.Start(ctx)
		<-ready

		// Act

		// Wait for refresh interval to start
		<-time.After(2 * time.Second)

		// Create a timeout context
		getTokenctx, cancelGetToken := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelGetToken()

		// Get a token while TokenGetter is refreshing
		token, err := lh.GetToken(getTokenctx)

		// Assert

		// Assert that the token is nil value
		if token != "" {
			t.Errorf("expected nil token value, got %s", token)
		}

		// Asser that the errror is the context error
		if getTokenctx.Err() != err {
			t.Errorf("expected context error %v to be equal to error returned from GetToken, got %v", getTokenctx.Err(), err)
		}
	})
}

func newPowerFlexClient(t *testing.T, addr string) *goscaleio.Client {
	client, err := goscaleio.NewClientWithArgs(addr, "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func newPowerFlexTestServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	return server
}
