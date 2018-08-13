package cmd

import (
	"io"
	"os"
	"fmt"
	"bufio"
	"strings"
	"github.com/spf13/cobra"
	"github.com/cyberark/secretless-broker/internal/app/secretlessctl/kubernetes"
)

var outputPath string

var injectCmd = &cobra.Command{
	Use: "inject [manifest file]",
	Short: "Injects a Secretless sidecar into a deployment manifest",
	Long: `Injects the Secretless sidecar to Pods, Deployments and ReplicaSets.
The sidecar will be injected into objects where a "secretlessConfig" annotation
references an existing ConfigMap. The ConfigMap should contain a valid
Secretless configuration.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var reader io.Reader

		if len(args) == 0 || args[0] == "-" {
			// Make sure there's a pipe to read from
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				cmd.Usage()
				os.Exit(1)
			}

			manifest := ""
			buf := make([]byte, 2048)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					break
				}
				manifest = manifest + string(buf[:n])
			}
			reader = strings.NewReader(manifest)
		} else {
			file, err := os.Open(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
			reader = bufio.NewReader(file)
		}

		var writer io.Writer
		if outputPath != "" {
			var err error
			writer, err = os.Create(outputPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		} else {
			writer = os.Stdout
		}

		err := kubernetes.InjectManifest(reader, writer)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	injectCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file")
}