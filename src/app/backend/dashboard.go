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

	"github.com/kubernetes/dashboard/src/app/backend/args"
	"github.com/kubernetes/dashboard/src/app/backend/auth"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	"github.com/kubernetes/dashboard/src/app/backend/auth/jwe"
	"github.com/kubernetes/dashboard/src/app/backend/cert"
	"github.com/kubernetes/dashboard/src/app/backend/cert/ecdsa"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	integrationapi "github.com/kubernetes/dashboard/src/app/backend/integration/api"
	"github.com/kubernetes/dashboard/src/app/backend/settings"
	"github.com/kubernetes/dashboard/src/app/backend/sync"
	"github.com/kubernetes/dashboard/src/app/backend/systembanner"
)

var (
	argInsecurePort              = pflag.Int("insecure-port", 9090, "port to listen to for incoming HTTP requests")
	argPort                      = pflag.Int("port", 8443, "secure port to listen to for incoming HTTPS requests")
	argInsecureBindAddress       = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --insecure-port, set to 127.0.0.1 for all interfaces")
	argBindAddress               = pflag.IP("bind-address", net.IPv4(0, 0, 0, 0), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	argDefaultCertDir            = pflag.String("default-cert-dir", "/certs", "directory path containing files from --tls-cert-file and --tls-key-file, used also when auto-generating certificates flag is set")
	argCertFile                  = pflag.String("tls-cert-file", "", "file containing the default x509 certificate for HTTPS")
	argKeyFile                   = pflag.String("tls-key-file", "", "file containing the default x509 private key matching --tls-cert-file")
	argApiserverHost             = pflag.String("apiserver-host", "", "address of the Kubernetes API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for local discovery attempt")
	argMetricsProvider           = pflag.String("metrics-provider", "sidecar", "select provider type for metrics, 'none' will not check metrics")
	argHeapsterHost              = pflag.String("heapster-host", "", "address of the Heapster API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for service proxy usage")
	argSidecarHost               = pflag.String("sidecar-host", "", "address of the Sidecar API server to connect to in the format of protocol://address:port, leave it empty if the binary runs inside cluster for service proxy usage")
	argKubeConfigFile            = pflag.String("kubeconfig", "", "path to kubeconfig file with authorization and master location information")
	argTokenTTL                  = pflag.Int("token-ttl", authApi.DefaultTokenTTL, "expiration time in seconds of JWE tokens generated by dashboard, set to 0 to avoid expiration")
	argAuthenticationMode        = pflag.StringSlice("authentication-mode", []string{authApi.Token.String()}, "comma-separated list of enabled authentication options, supports 'token' and 'basic' that should only be used if Kubernetes API server has --authorization-mode=ABAC and --basic-auth-file flags set")
	argMetricClientCheckPeriod   = pflag.Int("metric-client-check-period", 30, "time interval between separate metric client health checks in seconds")
	argAutoGenerateCertificates  = pflag.Bool("auto-generate-certificates", false, "enables automatic certificates generation used to serve HTTPS")
	argEnableInsecureLogin       = pflag.Bool("enable-insecure-login", false, "enables login view when the app is not served over HTTPS")
	argEnableSkip                = pflag.Bool("enable-skip-login", false, "enables skip button on the login page")
	argSystemBanner              = pflag.String("system-banner", "", "system banner message displayed in the app if non-empty, it accepts simple HTML")
	argSystemBannerSeverity      = pflag.String("system-banner-severity", "INFO", "severity of system banner, should be one of 'INFO', 'WARNING' or 'ERROR'")
	argAPILogLevel               = pflag.String("api-log-level", "INFO", "level of API request logging, should be one of 'NONE', 'INFO' or 'DEBUG'")
	argDisableSettingsAuthorizer = pflag.Bool("disable-settings-authorizer", false, "disables settings page user authorizer so anyone can access settings page")
	argNamespace                 = pflag.String("namespace", getEnv("POD_NAMESPACE", "kube-system"), "if non-default namespace is used encryption key will be created in the specified namespace")
	localeConfig                 = pflag.String("locale-config", "./locale_conf.json", "path to file containing the locale configuration")
)

func main() {
	// Set logging output to standard console out
	log.SetOutput(os.Stdout)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = flag.CommandLine.Parse(make([]string, 0)) // Init for glog calls in kubernetes packages

	// Initializes dashboard arguments holder so we can read them in other packages
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

	clientManager := client.NewClientManager(args.Holder.GetKubeConfigFile(), args.Holder.GetApiServerHost())
	versionInfo, err := clientManager.InsecureClient().Discovery().ServerVersion()
	if err != nil {
		handleFatalInitError(err)
	}

	log.Printf("Successful initial request to the apiserver, version: %s", versionInfo.String())

	// Init auth manager
	authManager := initAuthManager(clientManager)

	// Init settings manager
	settingsManager := settings.NewSettingsManager()

	// Init system banner manager
	systemBannerManager := systembanner.NewSystemBannerManager(args.Holder.GetSystemBanner(),
		args.Holder.GetSystemBannerSeverity())

	// Init integrations
	integrationManager := integration.NewIntegrationManager(clientManager)

	switch metricsProvider := args.Holder.GetMetricsProvider(); metricsProvider {
	case "sidecar":
		integrationManager.Metric().ConfigureSidecar(args.Holder.GetSidecarHost()).
			EnableWithRetry(integrationapi.SidecarIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))
	case "heapster":
		integrationManager.Metric().ConfigureHeapster(args.Holder.GetHeapsterHost()).
			EnableWithRetry(integrationapi.HeapsterIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))
	case "none":
		log.Print("no metrics provider selected, will not check metrics.")
	default:
		log.Printf("Invalid metrics provider selected: %s", metricsProvider)
		log.Print("Defaulting to use the Sidecar provider.")
		integrationManager.Metric().ConfigureSidecar(args.Holder.GetSidecarHost()).
			EnableWithRetry(integrationapi.SidecarIntegrationID, time.Duration(args.Holder.GetMetricClientCheckPeriod()))
	}

	apiHandler, err := handler.CreateHTTPAPIHandler(
		integrationManager,
		clientManager,
		authManager,
		settingsManager,
		systemBannerManager)
	if err != nil {
		handleFatalInitError(err)
	}

	var servingCerts []tls.Certificate
	if args.Holder.GetAutoGenerateCertificates() {
		log.Println("Auto-generating certificates")
		certCreator := ecdsa.NewECDSACreator(args.Holder.GetKeyFile(), args.Holder.GetCertFile(), elliptic.P256())
		certManager := cert.NewCertManager(certCreator, args.Holder.GetDefaultCertDir())
		servingCert, err := certManager.GetCertificates()
		if err != nil {
			handleFatalInitServingCertError(err)
		}
		servingCerts = []tls.Certificate{servingCert}
	} else if args.Holder.GetCertFile() != "" && args.Holder.GetKeyFile() != "" {
		certFilePath := args.Holder.GetDefaultCertDir() + string(os.PathSeparator) + args.Holder.GetCertFile()
		keyFilePath := args.Holder.GetDefaultCertDir() + string(os.PathSeparator) + args.Holder.GetKeyFile()
		servingCert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
		if err != nil {
			handleFatalInitServingCertError(err)
		}
		servingCerts = []tls.Certificate{servingCert}
	}

	// Run a HTTP server that serves static public files from './public' and handles API calls.
	http.Handle("/", handler.MakeGzipHandler(handler.CreateLocaleHandler()))
	http.Handle("/api/", apiHandler)
	http.Handle("/config", handler.AppHandler(handler.ConfigHandler))
	http.Handle("/api/sockjs/", handler.CreateAttachHandler("/api/sockjs"))
	http.Handle("/metrics", promhttp.Handler())

	// Listen for http or https
	if servingCerts != nil {
		log.Printf("Serving securely on HTTPS port: %d", args.Holder.GetPort())
		secureAddr := fmt.Sprintf("%s:%d", args.Holder.GetBindAddress(), args.Holder.GetPort())
		server := &http.Server{
			Addr:    secureAddr,
			Handler: http.DefaultServeMux,
			TLSConfig: &tls.Config{
				Certificates: servingCerts,
				MinVersion:   tls.VersionTLS12,
			},
		}
		go func() { log.Fatal(server.ListenAndServeTLS("", "")) }()
	} else {
		log.Printf("Serving insecurely on HTTP port: %d", args.Holder.GetInsecurePort())
		addr := fmt.Sprintf("%s:%d", args.Holder.GetInsecureBindAddress(), args.Holder.GetInsecurePort())
		go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()
	}
	select {}
}

func initAuthManager(clientManager clientapi.ClientManager) authApi.AuthManager {
	insecureClient := clientManager.InsecureClient()

	// Init default encryption key synchronizer
	synchronizerManager := sync.NewSynchronizerManager(insecureClient)
	keySynchronizer := synchronizerManager.Secret(args.Holder.GetNamespace(), authApi.EncryptionKeyHolderName)

	// Register synchronizer. Overwatch will be responsible for restarting it in case of error.
	sync.Overwatch.RegisterSynchronizer(keySynchronizer, sync.AlwaysRestart)

	// Init encryption key holder and token manager
	keyHolder := jwe.NewRSAKeyHolder(keySynchronizer)
	tokenManager := jwe.NewJWETokenManager(keyHolder)
	tokenTTL := time.Duration(args.Holder.GetTokenTTL())
	if tokenTTL != authApi.DefaultTokenTTL {
		tokenManager.SetTokenTTL(tokenTTL)
	}

	// Set token manager for client manager.
	clientManager.SetTokenManager(tokenManager)
	authModes := authApi.ToAuthenticationModes(args.Holder.GetAuthenticationMode())
	if len(authModes) == 0 {
		authModes.Add(authApi.Token)
	}

	// UI logic dictates this should be the inverse of the cli option
	authenticationSkippable := args.Holder.GetEnableSkipLogin()

	return auth.NewAuthManager(clientManager, tokenManager, authModes, authenticationSkippable)
}

func initArgHolder() {
	builder := args.GetHolderBuilder()
	builder.SetInsecurePort(*argInsecurePort)
	builder.SetPort(*argPort)
	builder.SetTokenTTL(*argTokenTTL)
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
	builder.SetSystemBanner(*argSystemBanner)
	builder.SetSystemBannerSeverity(*argSystemBannerSeverity)
	builder.SetAPILogLevel(*argAPILogLevel)
	builder.SetAuthenticationMode(*argAuthenticationMode)
	builder.SetAutoGenerateCertificates(*argAutoGenerateCertificates)
	builder.SetEnableInsecureLogin(*argEnableInsecureLogin)
	builder.SetDisableSettingsAuthorizer(*argDisableSettingsAuthorizer)
	builder.SetEnableSkipLogin(*argEnableSkip)
	builder.SetNamespace(*argNamespace)
	builder.SetLocaleConfig(*localeConfig)
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
