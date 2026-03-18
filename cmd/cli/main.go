package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(platformerv1.AddToScheme(scheme))
}

func main() {
	root := &cobra.Command{
		Use:   "platform",
		Short: "PlatFormer CLI — deploy serverless apps to Kubernetes",
		Long: `platform is the CLI for PlatFormer.
It reads and writes ServerlessApp CRDs in your current Kubernetes context.`,
	}

	// Flags inherited by all subcommands.
	root.PersistentFlags().StringP("namespace", "n", "default", "Kubernetes namespace")
	root.PersistentFlags().String("kubeconfig", "", "Path to kubeconfig (defaults to in-cluster or $KUBECONFIG)")

	// Register subcommands.
	root.AddCommand(
		newInitCmd(),
		newDeployCmd(scheme),
		newStatusCmd(scheme),
		newDestroyCmd(scheme),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// buildClient constructs a controller-runtime client from the current kubeconfig.
func buildClient(scheme *runtime.Scheme) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("kubeconfig: %w", err)
	}
	return client.New(cfg, client.Options{Scheme: scheme})
}
