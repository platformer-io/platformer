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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

func (p *AWSProvider) CreateLogGroup(ctx context.Context, name string) error {
	_, err := p.cloudwatchClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})
	if err != nil {
		// Ignore if already exists — idempotent on re-deploy after a partial failure.
		var alreadyExists *cloudwatchtypes.ResourceAlreadyExistsException
		if errors.As(err, &alreadyExists) {
			return nil
		}
		return fmt.Errorf("aws: create log group %q: %w", name, err)
	}
	return nil
}

func (p *AWSProvider) DeleteLogGroup(ctx context.Context, name string) error {
	_, err := p.cloudwatchClient.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: delete log group %q: %w", name, err)
	}
	return nil
}
