package workload

import (
	"fmt"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/jpeach/churn/pkg/k8s"
)

const LabelSelector = "app.kubernetes.io/managed-by=churn"

// DefaultDeleteResources is the default list of resources to
// consider as deletion candidates.
var DefaultDeleteCandidates = []schema.GroupVersionResource{
	{Group: "projectcontour.io", Version: "v1", Resource: "httpproxies"},
	{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
	{Group: "", Version: "v1", Resource: "services"},
	{Group: "", Version: "v1", Resource: "pods"},
}

func NewDeleterForPolicy(p Policy) (Task, error) {
	c, err := k8s.NewDynamicClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	f, err := k8s.NewDynamicInformerFactory(func(o *metav1.ListOptions) {
		// Only select objects with our matching label.
		o.LabelSelector = LabelSelector
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create informer: %w", err)
	}

	d := &resourceDeleter{
		client:    c,
		limit:     p.Limit,
		interval:  p.Interval,
		informers: make(map[schema.GroupVersionResource]informers.GenericInformer),
		stopChan:  make(chan struct{}),
	}

	for _, r := range p.Resources {
		if _, ok := d.informers[r]; ok {
			return nil, fmt.Errorf("duplicate resource: %s", r)
		}

		d.informers[r] = f.ForResource(r)
	}

	for _, i := range d.informers {
		i.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{})
		go i.Informer().Run(d.stopChan)
	}

	f.Start(d.stopChan)
	return d, nil
}

var _ Task = &resourceDeleter{}

type resourceDeleter struct {
	client    dynamic.Interface
	informers map[schema.GroupVersionResource]informers.GenericInformer

	limit    int
	interval time.Duration

	stopChan chan struct{}
}

func (d resourceDeleter) Interval() time.Duration {
	return d.interval
}

func (d resourceDeleter) Finalize() error {
	close(d.stopChan)
	return nil
}

func (d resourceDeleter) Perform() error {
	type candidate struct {
		gvr schema.GroupVersionResource
		obj *unstructured.Unstructured
	}

	candidates := make([]candidate, 0, 100)

	for gvr, i := range d.informers {
		if !i.Informer().HasSynced() {
			continue
		}

		objects, err := i.Lister().List(labels.Everything())
		if err != nil {
			return err
		}

		for _, o := range objects {
			candidates = append(candidates, candidate{
				gvr: gvr,
				obj: o.(*unstructured.Unstructured),
			})
		}
	}

	rand.Shuffle(len(candidates), func(i int, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	for i, c := range candidates {
		if i >= d.limit {
			break
		}

		err := d.client.Resource(c.gvr).Namespace(c.obj.GetNamespace()).Delete(
			c.obj.GetName(),
			&metav1.DeleteOptions{})
		if err != nil {
			return err
		}

		// TODO(jpeach): Wrap this in a package and emit with timestamp ...
		fmt.Printf("💀 deleted %s %s/%s\n", c.obj.GetKind(), c.obj.GetNamespace(), c.obj.GetName())
	}

	return nil
}
