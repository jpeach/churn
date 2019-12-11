package k8s

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ParseGroupVersionKind parses the GroupVersion and Kind strings
// from an APIResourceList and returns the corresponding GroupVersionKind.
func ParseGroupVersionKind(groupVersion string, r v1.APIResource) schema.GroupVersionKind {
	gv, err := schema.ParseGroupVersion(groupVersion)
	if err != nil {
		panic(err.Error())
	}

	return gv.WithKind(r.Kind)
}

func ParseGroupVersionResource(groupVersion string, r v1.APIResource) schema.GroupVersionResource {
	gv, err := schema.ParseGroupVersion(groupVersion)
	if err != nil {
		panic(err.Error())
	}

	return gv.WithResource(r.Name)
}
