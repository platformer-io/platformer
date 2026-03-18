package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
)

func newStatusCmd(scheme *runtime.Scheme) *cobra.Command {
	return &cobra.Command{
		Use:   "status <app-name>",
		Short: "Show the current status of a ServerlessApp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ns, _ := cmd.Flags().GetString("namespace")

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

			s := app.Status

			// Phase with colour indicator.
			phaseIcon := phaseIcon(s.Phase)
			fmt.Printf("\n%s  %s/%s  [%s]\n\n", phaseIcon, ns, name, s.Phase)

			printField("Function ARN    ", s.FunctionARN)
			printField("Function Version", s.FunctionVersion)
			printField("Execution Role  ", s.ExecutionRoleARN)
			printField("API Endpoint    ", s.APIEndpoint)
			printField("Log Group       ", s.LogGroupName)

			if len(s.Conditions) > 0 {
				fmt.Println("\nConditions:")
				for _, c := range s.Conditions {
					fmt.Printf("  %-20s %s  %s\n", c.Type, c.Status, c.Message)
				}
			}
			fmt.Println()
			return nil
		},
	}
}

func printField(label, value string) {
	if value == "" {
		return
	}
	fmt.Printf("  %-20s %s\n", label, value)
}

func phaseIcon(phase string) string {
	switch phase {
	case "Ready":
		return "✓"
	case "Provisioning":
		return "⟳"
	case "Failed":
		return "✗"
	case "Deleting":
		return "⊘"
	default:
		return "·"
	}
}
