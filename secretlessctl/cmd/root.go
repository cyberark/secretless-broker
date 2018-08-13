package cmd

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
  Use:   "secretlessctl",
  Short: "secretlessctl controls secretless deployments in Kubernetes",
  Long: "Find more information at https://secretless.io",
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
	rootCmd.AddCommand(injectCmd)
}