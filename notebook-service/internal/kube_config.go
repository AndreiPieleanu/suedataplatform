package internal

import (
	"flag"
	"path/filepath"
	"sync"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var kubeconfig *string
var kubeconfigOnce sync.Once // Ensures the flag is defined only once

var GetKubeConfig = func() (*rest.Config, error) {
	// Use sync.Once to ensure flag is only defined once
	kubeconfigOnce.Do(func() {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse() // Parse flags only once
	})

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	return config, nil
}
