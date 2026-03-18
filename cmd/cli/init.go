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
  memoryMB: 512
  timeoutSecs: 30
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

	cmd := &cobra.Command{
		Use:   "init <app-name>",
		Short: "Scaffold a new ServerlessApp manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			ns, _ := cmd.Flags().GetString("namespace")

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

	return cmd
}
