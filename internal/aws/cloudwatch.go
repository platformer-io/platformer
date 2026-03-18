package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func (p *AWSProvider) CreateLogGroup(ctx context.Context, name string) error {
	_, err := p.cloudwatchClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(name),
	})
	if err != nil {
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
