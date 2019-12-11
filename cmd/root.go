package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/jpeach/churn/pkg/k8s"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "churn",
	Short:        "churn a set of Kubernetes objects in a cluster",
	RunE:         Churn,
	SilenceUsage: true,
}

// Churn runs churn tasks ...
func Churn(cmd *cobra.Command, args []string) error {
	_, err := k8s.NewClientset()
	if err != nil {
		return err
	}

	return errors.New("not implemented")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("churn: %s\n", err)
		os.Exit(1)
	}
}
