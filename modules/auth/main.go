package main

import (
	"os"

	"k8s.io/klog/v2"

	"k8s.io/dashboard/auth/pkg/args"
	"k8s.io/dashboard/auth/pkg/environment"
	"k8s.io/dashboard/auth/pkg/router"
	"k8s.io/dashboard/client"

	// Importing route packages forces route registration
	_ "k8s.io/dashboard/auth/pkg/login"
)

func main() {
	client.Init(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KubeconfigPath()),
	)

	klog.V(1).InfoS("Listening and serving on", "address", args.Address())
	if err := router.Router().Run(args.Address()); err != nil {
		klog.ErrorS(err, "Router error")
		os.Exit(1)
	}
}
