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
	"os"

	"k8s.io/dashboard/certificates"
	"k8s.io/dashboard/certificates/ecdsa"
	"k8s.io/dashboard/client"
	"k8s.io/dashboard/web/pkg/args"
	"k8s.io/dashboard/web/pkg/environment"
	"k8s.io/dashboard/web/pkg/router"
	"k8s.io/klog/v2"

	// Importing route packages forces route registration
	_ "k8s.io/dashboard/web/pkg/config"
	_ "k8s.io/dashboard/web/pkg/locale"
	_ "k8s.io/dashboard/web/pkg/settings"
	_ "k8s.io/dashboard/web/pkg/systembanner"
)

func main() {
	client.Init(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KubeconfigPath()),
	)

	certCreator := ecdsa.NewECDSACreator(args.KeyFile(), args.CertFile(), elliptic.P256())
	certManager := certificates.NewCertManager(certCreator, args.DefaultCertDir(), args.AutoGenerateCertificates())
	certPath, keyPath, err := certManager.GetCertificatePaths()
	if err != nil {
		klog.Fatalf("Error while loading dashboard server certificates. Reason: %s", err)
	}

	if len(certPath) != 0 && len(keyPath) != 0 {
		serveTLS(certPath, keyPath)
	} else {
		serve()
	}
}

func serve() {
	klog.Infof("Serving insecurely on HTTP port: %d", args.InsecurePort())

	klog.V(1).InfoS("Listening and serving on", "address", args.InsecureAddress())
	if err := router.Router().Run(args.InsecureAddress()); err != nil {
		klog.ErrorS(err, "Router error")
		os.Exit(1)
	}
}

func serveTLS(certPath, keyPath string) {
	klog.Infof("Serving securely on HTTPS port: %d", args.Port())

	klog.V(1).InfoS("Listening and serving on", "address", args.Address())
	if err := router.Router().RunTLS(args.Address(), certPath, keyPath); err != nil {
		klog.ErrorS(err, "Router error")
		os.Exit(1)
	}
}
