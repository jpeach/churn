package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jpeach/churn/pkg/workload"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/jpeach/churn/pkg/k8s"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "churn",
	Short:         "churn a set of Kubernetes objects in a cluster",
	RunE:          Churn,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func parsePolicySpecs(args []string) ([]workload.PolicySpec, error) {
	var specs []workload.PolicySpec

	for _, arg := range args {
		spec, err := workload.ParsePolicySpec(arg)
		if err != nil {
			return nil, err
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

func newTaskForSpec(spec workload.PolicySpec, d k8s.DiscoveryHelper) (workload.Task, error) {
	p := workload.Policy{
		Interval: 60 * time.Second,
		Limit:    1,
	}

	for _, name := range spec.ResourceNames {
		gvr, _, err := d.ResourceFor(schema.GroupVersionResource{Resource: name})
		if err != nil {
			return nil, err
		}

		p.Resources = append(p.Resources, gvr)
	}

	for k, v := range spec.Parameters {
		switch k {
		case "interval":
			p.Interval, _ = time.ParseDuration(v)
		case "limit":
			p.Limit, _ = strconv.Atoi(v)
		default:
			return nil, fmt.Errorf("invalid parameter name '%s'", k)
		}
	}

	switch spec.Operation {
	case "delete":
		if len(p.Resources) == 0 {
			p.Resources = workload.DefaultDeleteCandidates
		}

		return workload.NewDeleterForPolicy(p)
	default:
		return nil, fmt.Errorf("invalid operation name '%s'", spec.Operation)
	}
}

// Churn runs churn tasks ...
func Churn(cmd *cobra.Command, args []string) error {
	specs, err := parsePolicySpecs(args)
	if err != nil {
		return err
	}

	if len(specs) == 0 {
		log.Printf("usage: churn TASKSPEC [TASKSPEC...]")
		return nil
	}

	cs, err := k8s.NewClientset()
	if err != nil {
		return err
	}

	d, err := k8s.NewDiscoveryHelper(cs)
	if err != nil {
		return err
	}

	tasks := []workload.Task{}

	for _, spec := range specs {
		task, err := newTaskForSpec(spec, d)
		if err != nil {
			return err
		}

		tasks = append(tasks, task)
	}

	stopChan := make(chan struct{})
	return workload.Run(stopChan, tasks)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("churn: %s\n", err)
		os.Exit(1)
	}
}
