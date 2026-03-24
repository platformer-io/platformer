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
