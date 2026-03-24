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

package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/platformer-io/platformer/internal/provider"
)

func (p *AWSProvider) CreateFunction(ctx context.Context, spec provider.FunctionSpec) (*provider.FunctionResult, error) {
	output, err := p.lambdaClient.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(spec.Name),
		Runtime:      types.Runtime(spec.Runtime),
		Handler:      aws.String("index.handler"),
		MemorySize:   aws.Int32(int32(spec.MemoryMB)),
		Timeout:      aws.Int32(int32(spec.TimeoutSecs)),
		Role:         aws.String(spec.ExecutionRole),
		Environment: &types.Environment{
			Variables: spec.Environment,
		},
		Code: &types.FunctionCode{
			S3Bucket: aws.String(spec.CodeBucket),
			S3Key:    aws.String(spec.CodeKey),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("aws: create function %q: %w", spec.Name, err)
	}

	// Wait for the function to become active before returning.
	// This prevents API Gateway wiring from failing on a function that isn't ready yet.
	waiter := lambda.NewFunctionActiveV2Waiter(p.lambdaClient)
	if err := waiter.Wait(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(spec.Name),
	}, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("aws: waiting for function %q to become active: %w", spec.Name, err)
	}

	return &provider.FunctionResult{
		ID:      aws.ToString(output.FunctionArn),
		Version: aws.ToString(output.Version),
	}, nil
}

func (p *AWSProvider) UpdateFunction(ctx context.Context, spec provider.FunctionSpec) (*provider.FunctionResult, error) {
	// Update code first, then configuration separately (Lambda API requires two calls).
	_, err := p.lambdaClient.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(spec.Name),
		S3Bucket:     aws.String(spec.CodeBucket),
		S3Key:        aws.String(spec.CodeKey),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: update function code %q: %w", spec.Name, err)
	}

	output, err := p.lambdaClient.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(spec.Name),
		Runtime:      types.Runtime(spec.Runtime),
		Handler:      aws.String("index.handler"),
		MemorySize:   aws.Int32(int32(spec.MemoryMB)),
		Timeout:      aws.Int32(int32(spec.TimeoutSecs)),
		Role:         aws.String(spec.ExecutionRole),
		Environment: &types.Environment{
			Variables: spec.Environment,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("aws: update function config %q: %w", spec.Name, err)
	}

	return &provider.FunctionResult{
		ID:      aws.ToString(output.FunctionArn),
		Version: aws.ToString(output.Version),
	}, nil
}

func (p *AWSProvider) DeleteFunction(ctx context.Context, name string) error {
	_, err := p.lambdaClient.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: delete function %q: %w", name, err)
	}
	return nil
}

func (p *AWSProvider) GetFunction(ctx context.Context, name string) (*provider.FunctionResult, error) {
	output, err := p.lambdaClient.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: get function %q: %w", name, err)
	}

	return &provider.FunctionResult{
		ID:      aws.ToString(output.Configuration.FunctionArn),
		Version: aws.ToString(output.Configuration.Version),
	}, nil
}
