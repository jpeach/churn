package cmd

import (
	"fmt"
	"log"
	"os"
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

// Churn runs churn tasks ...
func Churn(cmd *cobra.Command, args []string) error {
	cs, err := k8s.NewClientset()
	if err != nil {
		return err
	}

	resourceTypes := make([]schema.GroupVersionResource, 0, 100)

	d, err := k8s.NewDiscoveryHelper(cs)
	if err != nil {
		return err
	}

	// TODO(jpeach): only pull the cluster resource types if
	// we are actually going to need it.
	for _, group := range d.Resources() {
		for _, res := range group.APIResources {
			gvr := k8s.ParseGroupVersionResource(group.GroupVersion, res)
			resourceTypes = append(resourceTypes, k8s.ParseGroupVersionResource(group.GroupVersion, res))
		}
	}

	// TODO: add delete workload definition syntax:
	//	delete:all,httpproxies,max=10,rate=1

	deleter, err := workload.NewDeleter(workload.DefaultDeleteCandidates)
	if err != nil {
		return err
	}

	for {
		if err := deleter.DeleteObjects(10); err != nil {
			log.Printf("deletion error: %w", err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("churn: %s\n", err)
		os.Exit(1)
	}
}
