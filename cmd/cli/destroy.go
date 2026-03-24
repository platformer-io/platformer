// Copyright 2026 PlatFormer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

			ctx := context.Background()
			app := &platformerv1.ServerlessApp{}
			if err := k8s.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, app); err != nil {
				return fmt.Errorf("get ServerlessApp %s/%s: %w", ns, name, err)
			}

			if err := k8s.Delete(ctx, app); err != nil {
				return fmt.Errorf("delete ServerlessApp: %w", err)
			}

			fmt.Printf("🗑  Destroying %s...\n", name)

			// Poll until the resource is fully removed (finalizer runs cleanup).
			start := time.Now()
			pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			key := client.ObjectKey{Name: name, Namespace: ns}
			for {
				select {
				case <-pollCtx.Done():
					return fmt.Errorf("timed out waiting for %s to be destroyed", name)
				case <-time.After(2 * time.Second):
				}

				check := &platformerv1.ServerlessApp{}
				err := k8s.Get(pollCtx, key, check)
				if k8serrors.IsNotFound(err) {
					elapsed := time.Since(start).Round(time.Second)
					fmt.Printf("✔ Destroyed in %s\n", elapsed)
					return nil
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}
