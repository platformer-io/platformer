package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/platformer-io/platformer/internal/provider"
)

func (p *AWSProvider) CreateDatabase(ctx context.Context, spec provider.DatabaseSpec) (*provider.DatabaseResult, error) {
	results := &provider.DatabaseResult{
		ProviderMeta: make(map[string]string),
	}

	for _, table := range spec.Tables {
		output, err := p.dynamoDBClient.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName: aws.String(table.Name),
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("pk"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("pk"),
					KeyType:       types.KeyTypeHash,
				},
			},
			BillingMode: types.BillingModePayPerRequest,
		})
		if err != nil {
			return nil, fmt.Errorf("aws: create table %q: %w", table.Name, err)
		}
		results.ProviderMeta[table.Name] = aws.ToString(output.TableDescription.TableArn)
	}

	results.ID = spec.Name
	return results, nil
}

func (p *AWSProvider) DeleteDatabase(ctx context.Context, name string) error {
	// name here maps to a logical group; caller is responsible for passing individual table names.
	_, err := p.dynamoDBClient.DeleteTable(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: delete table %q: %w", name, err)
	}
	return nil
}

func (p *AWSProvider) GetDatabase(ctx context.Context, name string) (*provider.DatabaseResult, error) {
	output, err := p.dynamoDBClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: describe table %q: %w", name, err)
	}

	return &provider.DatabaseResult{
		ID: aws.ToString(output.Table.TableArn),
		ProviderMeta: map[string]string{
			"status": string(output.Table.TableStatus),
		},
	}, nil
}
