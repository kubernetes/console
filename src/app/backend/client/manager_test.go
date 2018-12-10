// Copyright 2017 The Kubernetes Authors.
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

package client

import (
	"crypto/tls"
	"net/http"
	"testing"

	restful "github.com/emicklei/go-restful"

  "github.com/kubernetes/dashboard/src/app/backend/args"
)

func TestNewClientManager(t *testing.T) {
	cases := []struct {
		kubeConfigPath, apiserverHost string
	}{
		{"", "test"},
	}

	for _, c := range cases {
		manager := NewClientManager(c.kubeConfigPath, c.apiserverHost)

		if manager == nil {
			t.Fatalf("NewClientManager(%s, %s): Expected manager not to be nil",
				c.kubeConfigPath, c.apiserverHost)
		}
	}
}

func TestClient(t *testing.T) {
  args.GetHolderBuilder().SetEnableSkipLogin(true)
	cases := []struct {
		request *restful.Request
	}{
		{
			&restful.Request{
				Request: &http.Request{
					Header: http.Header(map[string][]string{}),
				},
			},
		},
	}

	for _, c := range cases {
		manager := NewClientManager("", "http://localhost:8080")
		_, err := manager.Client(c.request)

		if err != nil {
			t.Fatalf("Client(%v): Expected client to be created but error was thrown:"+
				" %s", c.request, err.Error())
		}
	}
}

func TestCSRFKey(t *testing.T) {
	manager := NewClientManager("", "http://localhost:8080")
	key := manager.CSRFKey()

	if len(key) == 0 {
		t.Fatal("CSRFKey(): Expected csrf key to be autogenerated.")
	}
}

func TestConfig(t *testing.T) {
	cases := []struct {
		request  *restful.Request
		expected string
	}{
		{
			&restful.Request{
				Request: &http.Request{
					Header: http.Header(map[string][]string{
						"Authorization": {"Bearer test-token"},
					}),
					TLS: &tls.ConnectionState{},
				},
			},
			"test-token",
		},
	}

	for _, c := range cases {
		manager := NewClientManager("", "https://localhost:8080")
		cfg, err := manager.Config(c.request)

		if err != nil {
			t.Fatalf("Config(%v): Expected config to be created but error was thrown:"+
				" %s",
				c.request, err.Error())
		}

		if cfg.BearerToken != c.expected {
			t.Fatalf("Config(%v): Expected token to be %s but got %s",
				c.request, c.expected, cfg.BearerToken)
		}
	}
}

func TestClientCmdConfig(t *testing.T) {
  args.GetHolderBuilder().SetEnableSkipLogin(true)
	cases := []struct {
		request  *restful.Request
		expected string
	}{
		{
			&restful.Request{
				Request: &http.Request{
					Header: http.Header(map[string][]string{
						"Authorization": {"Bearer test-token"},
					}),
					TLS: &tls.ConnectionState{},
				},
			},
			"test-token",
		},
	}

	for _, c := range cases {
		manager := NewClientManager("", "https://localhost:8080")
		cmdCfg, err := manager.ClientCmdConfig(c.request)

		if err != nil {
			t.Fatalf("Config(%v): Expected client config to be created but error was thrown:"+
				" %s",
				c.request, err.Error())
		}

		var bearerToken string
		if cmdCfg != nil {
			cfg, err := cmdCfg.ClientConfig()
			if err != nil {
				t.Fatalf("Config(%v): Expected config to be created but error was thrown:"+
					" %s",
					c.request, err.Error())
			}
			bearerToken = cfg.BearerToken
		}

		if bearerToken != c.expected {
			t.Fatalf("Config(%v): Expected token to be %s but got %s",
				c.request, c.expected, bearerToken)
		}
	}
}

func TestVerberClient(t *testing.T) {
	manager := NewClientManager("", "http://localhost:8080")
	_, err := manager.VerberClient(&restful.Request{Request: &http.Request{TLS: &tls.ConnectionState{}}})

	if err != nil {
		t.Fatalf("VerberClient(): Expected verber client to be created but got error: %s",
			err.Error())
	}
}

func TestClientManager_InsecureClient(t *testing.T) {
	manager := NewClientManager("", "http://localhost:8080")
	if manager.InsecureClient() == nil {
		t.Fatalf("InsecureClient(): Expected insecure client not to be nil")
	}
}
