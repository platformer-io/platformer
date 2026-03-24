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
