package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/platformer-io/platformer/internal/provider"
)

// AWSProvider implements provider.CloudProvider using AWS SDK v2.
type AWSProvider struct {
	lambdaClient     *lambda.Client
	apiGatewayClient *apigatewayv2.Client
	iamClient        *iam.Client
	dynamoDBClient   *dynamodb.Client
	cloudwatchClient *cloudwatchlogs.Client
	region           string
}

// NewAWSProvider constructs an AWSProvider from a provider.Config.
// Credentials are resolved from the environment (AWS_ACCESS_KEY_ID,
// AWS_SECRET_ACCESS_KEY, AWS_PROFILE, or IAM role) via the default
// AWS credential chain.
func NewAWSProvider(cfg provider.Config) (*AWSProvider, error) {
	clients, err := newAWSClients(cfg.Region)
	if err != nil {
		return nil, err
	}
	return &AWSProvider{
		lambdaClient:     clients.lambda,
		apiGatewayClient: clients.apiGateway,
		iamClient:        clients.iam,
		dynamoDBClient:   clients.dynamoDB,
		cloudwatchClient: clients.cloudwatch,
		region:           cfg.Region,
	}, nil
}
