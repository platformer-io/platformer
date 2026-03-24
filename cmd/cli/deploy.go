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
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	platformerv1 "github.com/platformer-io/platformer/api/v1"
	appconfig "github.com/platformer-io/platformer/internal/config"
)

func newDeployCmd(scheme *runtime.Scheme) *cobra.Command {
	return &cobra.Command{
		Use:   "deploy <directory>",
		Short: "Deploy an app from a platformer.yaml directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			appDir := args[0]
			ns, _ := cmd.Flags().GetString("namespace")
			ctx := context.Background()

			// 1. Load and validate platformer.yaml.
			cfg, err := appconfig.Load(filepath.Join(appDir, "platformer.yaml"))
			if err != nil {
				return err
			}

			// 2. Verify handler file exists.
			handlerPath := filepath.Join(appDir, cfg.Handler)
			if _, err := os.Stat(handlerPath); err != nil {
				return fmt.Errorf("handler file not found: %s", handlerPath)
			}

			fmt.Printf("🚀 Deploying %s...\n", cfg.Name)

			// 3. Load AWS config.
			awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("aws config: %w", err)
			}
			region := awsCfg.Region
			if region == "" {
				region = "us-east-1"
			}

			// 4. Get AWS account ID.
			stsClient := sts.NewFromConfig(awsCfg)
			identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
			if err != nil {
				return fmt.Errorf("aws sts get-caller-identity: %w", err)
			}
			accountID := aws.ToString(identity.Account)

			// 5. Zip the src/ directory.
			zipData, err := zipSrcDir(appDir)
			if err != nil {
				return err
			}

			// 6. Create S3 bucket if not exists.
			bucketName := fmt.Sprintf("platformer-%s-%s", accountID, region)
			s3Client := s3.NewFromConfig(awsCfg)
			created, err := ensureBucket(ctx, s3Client, bucketName, region)
			if err != nil {
				return fmt.Errorf("s3 bucket: %w", err)
			}
			if created {
				fmt.Printf("✔ Created S3 bucket: %s\n", bucketName)
			} else {
				fmt.Printf("✔ Using S3 bucket: %s\n", bucketName)
			}

			// 7. Upload zip to S3.
			s3Key := fmt.Sprintf("apps/%s/function.zip", cfg.Name)
			size := int64(len(zipData))
			_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket:        aws.String(bucketName),
				Key:           aws.String(s3Key),
				Body:          bytes.NewReader(zipData),
				ContentLength: aws.Int64(size),
			})
			if err != nil {
				return fmt.Errorf("s3 upload: %w", err)
			}
			fmt.Printf("✔ Uploaded function code (%s)\n", formatBytes(size))

			// 8. Build ServerlessApp in memory and apply to cluster.
			app := buildServerlessApp(cfg, ns, bucketName, s3Key)
			k8s, err := buildClient(scheme)
			if err != nil {
				return err
			}

			existing := &platformerv1.ServerlessApp{}
			err = k8s.Get(ctx, client.ObjectKeyFromObject(app), existing)
			if err != nil {
				if err := k8s.Create(ctx, app); err != nil {
					return fmt.Errorf("create ServerlessApp: %w", err)
				}
			} else {
				existing.Spec = app.Spec
				if err := k8s.Update(ctx, existing); err != nil {
					return fmt.Errorf("update ServerlessApp: %w", err)
				}
				app = existing
			}
			fmt.Printf("✔ Applied ServerlessApp to cluster\n")

			// 9. Poll until Ready or Failed (5 min timeout).
			fmt.Printf("✔ Provisioning... (this takes ~20-30 seconds)\n")
			start := time.Now()
			pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			key := client.ObjectKeyFromObject(app)
			for {
				select {
				case <-pollCtx.Done():
					return fmt.Errorf("timed out waiting for %s to become Ready", cfg.Name)
				case <-time.After(2 * time.Second):
				}

				latest := &platformerv1.ServerlessApp{}
				if err := k8s.Get(pollCtx, key, latest); err != nil {
					continue
				}

				switch latest.Status.Phase {
				case "Ready":
					elapsed := time.Since(start).Round(time.Second)
					fmt.Printf("✔ Ready in %s\n", elapsed)
					fmt.Printf("\n🌐 Endpoint: %s\n", latest.Status.APIEndpoint)
					fmt.Printf("\nTest it:\n  curl %s\n", latest.Status.APIEndpoint)
					fmt.Printf("\nClean up:\n  platform destroy %s\n", cfg.Name)
					return nil
				case "Failed":
					fmt.Printf("✗ Deployment failed\n")
					for _, c := range latest.Status.Conditions {
						if c.Status == "False" {
							fmt.Printf("  %s: %s\n", c.Type, c.Message)
						}
					}
					return fmt.Errorf("deployment failed — run: kubectl describe serverlessapp %s -n %s", cfg.Name, ns)
				}
			}
		},
	}
}

// buildServerlessApp constructs a ServerlessApp from a PlatformerConfig.
func buildServerlessApp(cfg *appconfig.PlatformerConfig, ns, bucket, s3Key string) *platformerv1.ServerlessApp {
	app := &platformerv1.ServerlessApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: ns,
		},
		Spec: platformerv1.ServerlessAppSpec{
			Runtime:     cfg.Runtime,
			MemoryMB:    int32(cfg.Memory),
			TimeoutSecs: int32(cfg.Timeout),
			Code: platformerv1.CodeSource{
				S3Bucket: bucket,
				S3Key:    s3Key,
			},
			Environment: cfg.Env,
		},
	}

	if cfg.API != nil && cfg.API.Enabled {
		stage := cfg.API.Stage
		if stage == "" {
			stage = "prod"
		}
		app.Spec.API = &platformerv1.APIConfig{
			Enabled: true,
			Stage:   stage,
		}
	}

	if cfg.DB != nil {
		tables := make([]platformerv1.TableSpec, 0, len(cfg.DB.Tables))
		for _, t := range cfg.DB.Tables {
			tables = append(tables, platformerv1.TableSpec{Name: t.Name})
		}
		app.Spec.Database = &platformerv1.DatabaseConfig{Tables: tables}
	}

	return app
}

// zipSrcDir zips the contents of <appDir>/src/ with files at the zip root.
func zipSrcDir(appDir string) ([]byte, error) {
	srcDir := filepath.Join(appDir, "src")
	if _, err := os.Stat(srcDir); err != nil {
		return nil, fmt.Errorf("src/ directory not found in %s", appDir)
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(srcDir, path)
		rel = filepath.ToSlash(rel)
		f, err := w.Create(rel)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("zip src/: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}
	return buf.Bytes(), nil
}

// ensureBucket creates the bucket if it doesn't exist. Returns true if created.
func ensureBucket(ctx context.Context, client *s3.Client, bucket, region string) (bool, error) {
	input := &s3.CreateBucketInput{Bucket: aws.String(bucket)}
	if region != "us-east-1" {
		input.CreateBucketConfiguration = &s3types.CreateBucketConfiguration{
			LocationConstraint: s3types.BucketLocationConstraint(region),
		}
	}

	_, err := client.CreateBucket(ctx, input)
	if err == nil {
		return true, nil
	}

	var owned *s3types.BucketAlreadyOwnedByYou
	var exists *s3types.BucketAlreadyExists
	if errors.As(err, &owned) || errors.As(err, &exists) {
		return false, nil
	}
	return false, err
}

// lambdaHandler derives the Lambda handler string from a file path.
// "src/index.js" → "index.handler"
func lambdaHandler(handlerFile string) string {
	base := filepath.Base(handlerFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return name + ".handler"
}

func formatBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
	}
}
