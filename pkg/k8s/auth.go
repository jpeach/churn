package k8s

// Make sure we import the client-go auth provider plugins.

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)
