package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "churn",
	Short: "churn a set of Kubernetes objects in a cluster",
	Run:   func(cmd *cobra.Command, args []string) {},
}

// Churn runs churn tasks ...
func Churn(cmd *cobra.Command, args []string) {
	log.Printf("not yet ...")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
