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

package main

import (
	"crypto/elliptic"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"

	"k8s.io/dashboard/api/pkg/args"
	"k8s.io/dashboard/api/pkg/handler"
	"k8s.io/dashboard/api/pkg/integration"
	integrationapi "k8s.io/dashboard/api/pkg/integration/api"
	"k8s.io/dashboard/certificates"
	"k8s.io/dashboard/certificates/ecdsa"
	"k8s.io/dashboard/client"
)

var (
	argInsecurePort             = pflag.Int("insecure-port", 9000, "port to listen to for incoming HTTP requests")
	argPort                     = pflag.Int("port", 9001, "secure port to listen to for incoming HTTPS requests")
	argInsecureBindAddress      = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --insecure-port, set to 127.0.0.1 for all interfaces")
	argBindAddress              = pflag.IP("bind-address", net.IPv4(0, 0, 0, 0), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	argDefaultCertDir           = pflag.String("default-cert-dir", "/certs", "directory path containing files from --tls-cert-file and --tls-key-file, used also when auto-generating certificates flag is set")
	argCertFile                 = pflag.String("tls-cert-file", "", "file containing the default x509 certificate for HTTPS")
	argKeyFile                  = pflag.String("tls-key-file", "", "file containing the default x509 private key matching --tls-cert-file")
	argApiserverHost            = pflag.String("apiserver-host", "", "address of the Kubernetes API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for local discovery attempt")
	argApiserverSkipTLSVerify   = pflag.Bool("apiserver-skip-tls-verify", false, "enable if connection with remote Kubernetes API server should skip TLS verify")
	argMetricsProvider          = pflag.String("metrics-provider", "sidecar", "select provider type for metrics, 'none' will not check metrics")
	argHeapsterHost             = pflag.String("heapster-host", "", "address of the Heapster API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for service proxy usage")
	argSidecarHost              = pflag.String("sidecar-host", "", "address of the Sidecar API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for service proxy usage")
	argKubeConfigFile           = pflag.String("kubeconfig", "", "path to kubeconfig file with authorization and control plane location information")
	argMetricClientCheckPeriod  = pflag.Int("metric-client-check-period", 30, "time interval between separate metric client health checks in seconds")
	argAutoGenerateCertificates = pflag.Bool("auto-generate-certificates", false, "enables automatic certificates generation used to serve HTTPS")
	argEnableInsecureLogin      = pflag.Bool("enable-insecure-login", false, "enables login view when the app is not served over HTTPS")
	argEnableSkip               = pflag.Bool("enable-skip-login", false, "enables skip button on the login page")
	argAPILogLevel              = pflag.String("api-log-level", "INFO", "level of API request logging, should be one of 'NONE', 'INFO' or 'DEBUG'")
	argNamespace                = pflag.String("namespace", getEnv("POD_NAMESPACE", "kube-system"), "if non-default namespace is used encryption key will be created in the specified namespace")
)

func main() {
	// Set logging output to standard console out
	log.SetOutput(os.Stdout)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = flag.CommandLine.Parse(make([]string, 0)) // Init for glog calls in kubernetes packages

	// Initializes dashboard arguments holder, so we can read them in other packages
	initArgHolder()

	if args.Holder.GetApiServerHost() != "" {
		log.Printf("Using apiserver-host location: %s", args.Holder.GetApiServerHost())
	}
	if args.Holder.GetKubeConfigFile() != "" {
		log.Printf("Using kubeconfig file: %s", args.Holder.GetKubeConfigFile())
	}
	if args.Holder.GetNamespace() != "" {
		log.Printf("Using namespace: %s", args.Holder.GetNamespace())
	}

	client.Init(
		client.WithKubeconfig(args.Holder.GetKubeConfigFile()),
		client.WithMasterUrl(args.Holder.GetApiServerHost()),
		client.WithInsecureTLSSkipVerify(args.Holder.GetApiServerSkipTLSVerify()),
	)

	versionInfo, err := client.InClusterClient().Discovery().ServerVersion()
	if err != nil {
		handleFatalInitError(err)
	}

	log.Printf("Successful initial request to the apiserver, version: %s", versionInfo.String())

	// Init integrations
	integrationManager := integration.NewIntegrationManager()

	switch metricsProvider := args.Holder.GetMetricsProvider(); metricsProvider {
	case "sidecar":
		integrationManager.Metric().ConfigureSidecar(args.Holder.GetSidecarHost()).
			EnableWithRetry(integrationapi.SidecarIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))
	case "none":
		log.Print("no metrics provider selected, will not check metrics.")
	default:
		log.Printf("Invalid metrics provider selected: %s", metricsProvider)
		log.Print("Defaulting to use the Sidecar provider.")
		integrationManager.Metric().ConfigureSidecar(args.Holder.GetSidecarHost()).
			EnableWithRetry(integrationapi.SidecarIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))
	}

	apiHandler, err := handler.CreateHTTPAPIHandler(integrationManager)
	if err != nil {
		handleFatalInitError(err)
	}

	certCreator := ecdsa.NewECDSACreator(args.Holder.GetKeyFile(), args.Holder.GetCertFile(), elliptic.P256())
	certManager := certificates.NewCertManager(certCreator, args.Holder.GetDefaultCertDir(), args.Holder.GetAutoGenerateCertificates())
	certs, err := certManager.GetCertificates()
	if err != nil {
		handleFatalInitServingCertError(err)
	}

	http.Handle("/api/", apiHandler)
	http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	http.Handle("/metrics", promhttp.Handler())

	if certs != nil {
		serveTLS(certs)
	} else {
		serve()
	}

	select {}
}

func serve() {
	log.Printf("Serving insecurely on HTTP port: %d", args.Holder.GetInsecurePort())
	addr := fmt.Sprintf("%s:%d", args.Holder.GetInsecureBindAddress(), args.Holder.GetInsecurePort())
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()
}

func serveTLS(certificates []tls.Certificate) {
	log.Printf("Serving securely on HTTPS port: %d", args.Holder.GetPort())
	secureAddr := fmt.Sprintf("%s:%d", args.Holder.GetBindAddress(), args.Holder.GetPort())
	server := &http.Server{
		Addr:    secureAddr,
		Handler: http.DefaultServeMux,
		TLSConfig: &tls.Config{
			Certificates: certificates,
			MinVersion:   tls.VersionTLS12,
		},
	}
	go func() { log.Fatal(server.ListenAndServeTLS("", "")) }()
}

func initArgHolder() {
	builder := args.GetHolderBuilder()
	builder.SetInsecurePort(*argInsecurePort)
	builder.SetPort(*argPort)
	builder.SetMetricClientCheckPeriod(*argMetricClientCheckPeriod)
	builder.SetInsecureBindAddress(*argInsecureBindAddress)
	builder.SetBindAddress(*argBindAddress)
	builder.SetDefaultCertDir(*argDefaultCertDir)
	builder.SetCertFile(*argCertFile)
	builder.SetKeyFile(*argKeyFile)
	builder.SetApiServerHost(*argApiserverHost)
	builder.SetMetricsProvider(*argMetricsProvider)
	builder.SetHeapsterHost(*argHeapsterHost)
	builder.SetSidecarHost(*argSidecarHost)
	builder.SetKubeConfigFile(*argKubeConfigFile)
	builder.SetAPILogLevel(*argAPILogLevel)
	builder.SetAutoGenerateCertificates(*argAutoGenerateCertificates)
	builder.SetEnableInsecureLogin(*argEnableInsecureLogin)
	builder.SetEnableSkipLogin(*argEnableSkip)
	builder.SetNamespace(*argNamespace)
	builder.SetApiServerSkipTLSVerify(*argApiserverSkipTLSVerify)
}

/**
 * Handles fatal init error that prevents server from doing any work. Prints verbose error
 * message and quits the server.
 */
func handleFatalInitError(err error) {
	log.Fatalf("Error while initializing connection to Kubernetes apiserver. "+
		"This most likely means that the cluster is misconfigured (e.g., it has "+
		"invalid apiserver certificates or service account's configuration) or the "+
		"--apiserver-host param points to a server that does not exist. Reason: %s\n"+
		"Refer to our FAQ and wiki pages for more information: "+
		"https://github.com/kubernetes/dashboard/wiki/FAQ", err)
}

/**
 * Handles fatal init errors encountered during service cert loading.
 */
func handleFatalInitServingCertError(err error) {
	log.Fatalf("Error while loading dashboard server certificates. Reason: %s", err)
}

/**
* Lookup the environment variable provided and set to default value if variable isn't found
 */
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}
