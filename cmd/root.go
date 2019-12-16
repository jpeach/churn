package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jpeach/churn/pkg/workload"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/jpeach/churn/pkg/k8s"
	"github.com/spf13/cobra"
)

// MinimumInterval ...
const MinimumInterval = time.Second

// ErrUsage ...
func ErrUsage(cmd *cobra.Command) {
	fmt.Printf(cmd.UsageString())
	os.Exit(64)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "churn [flags] OPERATION[:RESOURCE|PARAM=VALUE[,RESOURCE|PARAM=VALUE]...]",
	RunE:          Churn,
	SilenceUsage:  true,
	SilenceErrors: true,

	Short: "churn a set of Kubernetes objects in a cluster",
	Long:  "churn generates Kubernetes object workloads in a cluster",
	Example: `  Delete 1 default resource each 5sec:
    $ churn delete:limit=1,interval=5s

  Delete pods and httpproxies at different rates:
    $ churn delete:pods,limit=10,interval=1m \
        delete:httpproxies,limit=1,interval=30s`,
}

func churnHelp(c *cobra.Command, _ []string) {
	fmt.Printf("%s\n", c.Long)
	fmt.Printf("\nUsage:\n  %s\n", c.UseLine())

	fmt.Printf(`
Operations:
  The general syntax of an operation specification is the name of the
  workload operation, separated from an optional comma-separated list
  of resources and parameters by a colon. A parameter is a key=value
  token that controls an aspect of the workload task.

Operation Types:
  delete    Deletes Kubernetes API objects. If no resources are
            specified, this operation deletes httpproxies, ingresses,
            services and pods.

Parameters:
  limit=COUNT
            Apply the operation oon up to COUNT objects per interval.
  interval=DURATION
            Apply the operation once per DURATION interval. The
            DURATION is a Go time.Duration string, e.g. "1m30s".
`)

	fmt.Printf("\nFlags:\n%s", c.LocalFlags().FlagUsages())

	if c.HasExample() {
		fmt.Printf("\nExamples:\n%s\n", c.Example)
	}
}

func init() {
	rootCmd.SetHelpFunc(churnHelp)
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

	if p.Interval < MinimumInterval {
		return nil, fmt.Errorf("interval is less than the minimum of %s", MinimumInterval)
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
		ErrUsage(cmd)
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
