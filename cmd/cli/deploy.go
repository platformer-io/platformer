package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
)

func newDeployCmd(scheme *runtime.Scheme) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy <manifest.yaml>",
		Short: "Apply a ServerlessApp manifest to the cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := args[0]
			ns, _ := cmd.Flags().GetString("namespace")

			// Read manifest.
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				return fmt.Errorf("read %s: %w", manifestPath, err)
			}

			// Decode into ServerlessApp.
			codec := serializer.NewCodecFactory(scheme)
			obj, _, err := codec.UniversalDeserializer().Decode(data, nil, nil)
			if err != nil {
				return fmt.Errorf("decode manifest: %w", err)
			}
			app, ok := obj.(*platformerv1.ServerlessApp)
			if !ok {
				return fmt.Errorf("manifest is not a ServerlessApp")
			}
			if app.Namespace == "" {
				app.Namespace = ns
			}

			k8s, err := buildClient(scheme)
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Create or update (server-side apply).
			existing := &platformerv1.ServerlessApp{}
			err = k8s.Get(ctx, client.ObjectKeyFromObject(app), existing)
			if err != nil {
				// Create new.
				if err := k8s.Create(ctx, app); err != nil {
					return fmt.Errorf("create ServerlessApp: %w", err)
				}
				fmt.Printf("✓ Created %s/%s\n", app.Namespace, app.Name)
			} else {
				// Update existing.
				existing.Spec = app.Spec
				if err := k8s.Update(ctx, existing); err != nil {
					return fmt.Errorf("update ServerlessApp: %w", err)
				}
				fmt.Printf("✓ Updated %s/%s\n", app.Namespace, app.Name)
			}

			fmt.Printf("  Track progress: platform status %s -n %s\n", app.Name, app.Namespace)
			return nil
		},
	}
}
