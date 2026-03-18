package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
)

func newDestroyCmd(scheme *runtime.Scheme) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "destroy <app-name>",
		Short: "Delete a ServerlessApp and all its cloud resources",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ns, _ := cmd.Flags().GetString("namespace")

			if !force {
				fmt.Printf("This will delete %s/%s and ALL its AWS resources.\n", ns, name)
				fmt.Print("Type the app name to confirm: ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(input) != name {
					fmt.Println("Aborted.")
					return nil
				}
			}

			k8s, err := buildClient(scheme)
			if err != nil {
				return err
			}

			app := &platformerv1.ServerlessApp{}
			if err := k8s.Get(context.Background(), client.ObjectKey{
				Name:      name,
				Namespace: ns,
			}, app); err != nil {
				return fmt.Errorf("get ServerlessApp %s/%s: %w", ns, name, err)
			}

			if err := k8s.Delete(context.Background(), app); err != nil {
				return fmt.Errorf("delete ServerlessApp: %w", err)
			}

			fmt.Printf("✓ Deletion triggered for %s/%s\n", ns, name)
			fmt.Printf("  The controller is now cleaning up AWS resources.\n")
			fmt.Printf("  Track progress: platform status %s -n %s\n", name, ns)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
