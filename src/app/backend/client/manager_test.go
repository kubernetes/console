package client

import (
	"github.com/emicklei/go-restful"
	"net/http"
	"testing"
)

func TestNewClientManager(t *testing.T) {
	cases := []struct {
		kubeConfigPath, apiserverHost string
	}{
		{"", ""},
		{"test", ""},
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
		{nil},
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
					Header: http.Header(map[string][]string{}),
				},
			},
			"",
		},
		{
			&restful.Request{
				Request: &http.Request{
					Header: http.Header(map[string][]string{
						"Authorization": []string{"Bearer test-token"},
					}),
				},
			},
			"test-token",
		},
		{nil, ""},
	}

	for _, c := range cases {
		manager := NewClientManager("", "http://localhost:8080")
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

func TestVerberClient(t *testing.T) {
	manager := NewClientManager("", "http://localhost:8080")
	_, err := manager.VerberClient(nil)

	if err != nil {
		t.Fatalf("VerberClient(): Expected verber client to be created but got error: %s",
			err.Error())
	}
}
