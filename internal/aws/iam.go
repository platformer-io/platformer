package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
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
		// If the role already exists (retry after partial failure), fetch and return it.
		var alreadyExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &alreadyExists) {
			existing, getErr := p.iamClient.GetRole(ctx, &iam.GetRoleInput{
				RoleName: aws.String(spec.Name),
			})
			if getErr != nil {
				return nil, fmt.Errorf("aws: role %q already exists but could not be fetched: %w", spec.Name, getErr)
			}
			return &provider.RoleResult{ID: aws.ToString(existing.Role.Arn)}, nil
		}
		return nil, fmt.Errorf("aws: create role %q: %w", spec.Name, err)
	}

	// Wait for IAM role to propagate before Lambda tries to assume it.
	time.Sleep(10 * time.Second)

	// Attach the AWS-managed basic Lambda execution policy (allows CloudWatch logging).
	_, err = p.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(spec.Name),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: attach basic execution policy to role %q: %w", spec.Name, err)
	}

	// Always attach an inline S3 read policy so Lambda can fetch its deployment package.
	s3ReadPolicy := `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["s3:GetObject"],
    "Resource": "arn:aws:s3:::platformer-*/*"
  }]
}`
	_, err = p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(spec.Name),
		PolicyName:     aws.String(spec.Name + "-s3-read"),
		PolicyDocument: aws.String(s3ReadPolicy),
	})
	if err != nil {
		return nil, fmt.Errorf("aws: attach s3 read policy to role %q: %w", spec.Name, err)
	}

	// If a DynamoDB table ARN is provided, attach a scoped inline policy (least-privilege).
	if spec.DatabaseARN != "" {
		dynamoPolicy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan"
    ],
    "Resource": "%s"
  }]
}`, spec.DatabaseARN)
		_, err = p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(spec.Name),
			PolicyName:     aws.String(spec.Name + "-dynamodb"),
			PolicyDocument: aws.String(dynamoPolicy),
		})
		if err != nil {
			return nil, fmt.Errorf("aws: attach dynamodb policy to role %q: %w", spec.Name, err)
		}
	}

	return &provider.RoleResult{
		ID: aws.ToString(role.Role.Arn),
	}, nil
}

func (p *AWSProvider) DeleteExecutionRole(ctx context.Context, name string) error {
	// Detach all managed policies before deleting the role (AWS requirement).
	attached, err := p.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("aws: list attached policies for role %q: %w", name, err)
	}
	for _, policy := range attached.AttachedPolicies {
		_, err := p.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(name),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("aws: detach policy %q from role %q: %w", aws.ToString(policy.PolicyArn), name, err)
		}
	}

	// Delete inline policies before deleting the role.
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
