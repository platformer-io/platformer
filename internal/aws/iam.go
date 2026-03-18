package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/platformer-io/platformer/internal/provider"
)

// lambdaAssumeRolePolicy allows Lambda to assume this execution role.
const lambdaAssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "Service": "lambda.amazonaws.com" },
    "Action": "sts:AssumeRole"
  }]
}`

func (p *AWSProvider) CreateExecutionRole(ctx context.Context, spec provider.RoleSpec) (*provider.RoleResult, error) {
	role, err := p.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(spec.Name),
		AssumeRolePolicyDocument: aws.String(lambdaAssumeRolePolicy),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: create role %q: %w", spec.Name, err)
	}

	// Attach each inline policy document.
	for i, doc := range spec.PolicyDocs {
		policyName := fmt.Sprintf("%s-policy-%d", spec.Name, i)
		_, err := p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(spec.Name),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(doc),
		})
		if err != nil {
			return nil, fmt.Errorf("aws: attach policy %d to role %q: %w", i, spec.Name, err)
		}
	}

	return &provider.RoleResult{
		ID: aws.ToString(role.Role.Arn),
	}, nil
}

func (p *AWSProvider) DeleteExecutionRole(ctx context.Context, name string) error {
	// Inline policies must be deleted before the role can be deleted.
	policies, err := p.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: list policies for role %q: %w", name, err)
	}
	for _, policyName := range policies.PolicyNames {
		_, err := p.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(name),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return fmt.Errorf("aws: delete policy %q from role %q: %w", policyName, name, err)
		}
	}

	_, err = p.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: delete role %q: %w", name, err)
	}
	return nil
}

func (p *AWSProvider) GetExecutionRole(ctx context.Context, name string) (*provider.RoleResult, error) {
	output, err := p.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: get role %q: %w", name, err)
	}

	return &provider.RoleResult{
		ID: aws.ToString(output.Role.Arn),
	}, nil
}
