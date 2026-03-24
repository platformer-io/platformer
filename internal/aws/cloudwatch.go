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
