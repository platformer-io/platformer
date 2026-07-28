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
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

const appTemplate = `apiVersion: platformer.io/v1
kind: ServerlessApp
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  runtime: {{ .Runtime }}
  tier: {{ .Tier }}
  code:
    s3Bucket: YOUR_BUCKET
    s3Key: {{ .Name }}/function.zip
  environment:
    APP_ENV: production
  api:
    enabled: true
    stage: prod
  # database:
  #   tables:
  #     - name: {{ .Name }}-data
`

func newInitCmd() *cobra.Command {
	var runtime string
	var tier string

	cmd := &cobra.Command{
		Use:   "init <app-name>",
		Short: "Scaffold a new ServerlessApp manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ns, _ := cmd.Flags().GetString("namespace")

			validTiers := map[string]bool{"small": true, "medium": true, "large": true, "custom": true}
			if !validTiers[tier] {
				return fmt.Errorf("--tier must be one of: small, medium, large, custom (got %q)", tier)
			}

			outFile := filepath.Join(".", name+".yaml")

			f, err := os.Create(outFile)
			if err != nil {
				return fmt.Errorf("create %s: %w", outFile, err)
			}
			defer f.Close()

			tmpl, err := template.New("app").Parse(appTemplate)
			if err != nil {
				return err
			}

			if err := tmpl.Execute(f, map[string]string{
				"Name":      name,
				"Namespace": ns,
				"Runtime":   runtime,
				"Tier":      tier,
			}); err != nil {
				return err
			}

			fmt.Printf("✓ Created %s\n", outFile)
			fmt.Printf("  Edit the manifest, then run: platform deploy %s\n", outFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&runtime, "runtime", "nodejs22.x",
		"Lambda runtime (nodejs22.x | python3.13 | provided.al2023)")
	cmd.Flags().StringVar(&tier, "tier", "medium",
		"Performance tier: small (128 MB/10 s) | medium (512 MB/30 s) | large (1024 MB/60 s) | custom")

	return cmd
}
