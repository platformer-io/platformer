package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/platformer-io/platformer/internal/provider"
)

func (p *AWSProvider) CreateAPIEndpoint(ctx context.Context, spec provider.APISpec) (*provider.APIResult, error) {
	api, err := p.apiGatewayClient.CreateApi(ctx, &apigatewayv2.CreateApiInput{
		Name:         aws.String(spec.Name),
		ProtocolType: types.ProtocolType(spec.Protocol),
		Target:       aws.String(spec.TargetID),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: create api %q: %w", spec.Name, err)
	}

	stage, err := p.apiGatewayClient.CreateStage(ctx, &apigatewayv2.CreateStageInput{
		ApiId:      api.ApiId,
		StageName:  aws.String(spec.Stage),
		AutoDeploy: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: create stage %q for api %q: %w", spec.Stage, spec.Name, err)
	}

	endpoint := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s",
		aws.ToString(api.ApiId), p.region, aws.ToString(stage.StageName))

	return &provider.APIResult{
		ID:       aws.ToString(api.ApiId),
		Endpoint: endpoint,
	}, nil
}

func (p *AWSProvider) UpdateAPIEndpoint(ctx context.Context, spec provider.APISpec) (*provider.APIResult, error) {
	output, err := p.apiGatewayClient.UpdateApi(ctx, &apigatewayv2.UpdateApiInput{
		ApiId:  aws.String(spec.TargetID),
		Name:   aws.String(spec.Name),
		Target: aws.String(spec.TargetID),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: update api %q: %w", spec.Name, err)
	}

	endpoint := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s",
		aws.ToString(output.ApiId), p.region, spec.Stage)

	return &provider.APIResult{
		ID:       aws.ToString(output.ApiId),
		Endpoint: endpoint,
	}, nil
}

func (p *AWSProvider) DeleteAPIEndpoint(ctx context.Context, id string) error {
	_, err := p.apiGatewayClient.DeleteApi(ctx, &apigatewayv2.DeleteApiInput{
		ApiId: aws.String(id),
	})
	if err != nil {
		return fmt.Errorf("aws: delete api %q: %w", id, err)
	}
	return nil
}

func (p *AWSProvider) GetAPIEndpoint(ctx context.Context, id string) (*provider.APIResult, error) {
	output, err := p.apiGatewayClient.GetApi(ctx, &apigatewayv2.GetApiInput{
		ApiId: aws.String(id),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: get api %q: %w", id, err)
	}

	return &provider.APIResult{
		ID:       aws.ToString(output.ApiId),
		Endpoint: aws.ToString(output.ApiEndpoint),
	}, nil
}
