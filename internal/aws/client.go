package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

type awsClients struct {
	lambda     *lambda.Client
	apiGateway *apigatewayv2.Client
	iam        *iam.Client
	dynamoDB   *dynamodb.Client
	cloudwatch *cloudwatchlogs.Client
}

// newAWSClients loads the default AWS credential chain and constructs
// all service clients for the given region.
func newAWSClients(region string) (*awsClients, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("aws: load config: %w", err)
	}

	return &awsClients{
		lambda:     lambda.NewFromConfig(cfg),
		apiGateway: apigatewayv2.NewFromConfig(cfg),
		iam:        iam.NewFromConfig(cfg),
		dynamoDB:   dynamodb.NewFromConfig(cfg),
		cloudwatch: cloudwatchlogs.NewFromConfig(cfg),
	}, nil
}
