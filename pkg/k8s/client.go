package k8s

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sirupsen/logrus/hooks/test"
	velero "github.com/vmware-tanzu/velero/pkg/discovery"
)

// UserAgent is the default value for the USer-Agent header.
var UserAgent = "churn"

// DefaultClientConfig() returns the default Kubernetes client config.
func DefaultClientConfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	return config.ClientConfig()
}

// NewClientset returns a new defaulted *kubernetes.Clientset.
func NewClientset() (*kubernetes.Clientset, error) {
	restConfig, err := DefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	return kubernetes.NewForConfig(restConfig)
}

// NewDynamicClientset returns a new defaulted dynamic.Interface.
func NewDynamicClientset() (dynamic.Interface, error) {
	restConfig, err := DefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	return dynamic.NewForConfig(restConfig)
}

type DiscoveryHelper = velero.Helper

// NewDiscoveryHelper returns a new resource discovery helper.
func NewDiscoveryHelper(cs *kubernetes.Clientset) (DiscoveryHelper, error) {
	logger, _ := test.NewNullLogger()

	return velero.NewHelper(memory.NewMemCacheClient(cs.Discovery()), logger)
}

// NewDynamicInformerFactory returns an configured dynamic informer factory.
func NewDynamicInformerFactory(optionsFunc dynamicinformer.TweakListOptionsFunc) (dynamicinformer.DynamicSharedInformerFactory, error) {
	restConfig, err := DefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	dynamicInterface, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	dynamicFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicInterface,
		10*time.Minute,
		v1.NamespaceAll,
		optionsFunc,
	)

	return dynamicFactory, nil
}

// NewDynamicInformer returns an informer for the given resource type.
func NewDynamicInformer(resource schema.GroupVersionResource, optionsFunc dynamicinformer.TweakListOptionsFunc) (informers.GenericInformer, error) {
	restConfig, err := DefaultClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build client config: %w", err)
	}

	dynamicInterface, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicInterface,
		10*time.Minute,
		v1.NamespaceAll,
		optionsFunc,
	)

	return informerFactory.ForResource(resource), nil
}
