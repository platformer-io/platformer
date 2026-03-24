# hello-world

A minimal Node.js Lambda function deployed via PlatFormer. Returns a JSON greeting with the request path and a timestamp.

## Prerequisites

- AWS CLI configured: `aws configure` (access key, secret, region: `us-east-1`)
- Verify credentials: `aws sts get-caller-identity`
- A running kind cluster with the PlatFormer CRD installed
- PlatFormer operator running: `go run cmd/operator/main.go`
- `platform` CLI built: `go build -o bin/platform ./cmd/cli`

## Deploy

```bash
platform deploy examples/hello-world/
```

`platform deploy` will automatically:

1. Read `platformer.yaml` from the directory
2. Zip the `src/` directory
3. Create an S3 bucket `platformer-<account-id>-<region>` (if it doesn't exist)
4. Upload the function zip to S3
5. Apply a `ServerlessApp` to your Kubernetes cluster
6. Poll until the function is live

Expected output:

```
🚀 Deploying hello-world...
✔ Created S3 bucket: platformer-123456789012-us-east-1
✔ Uploaded function code (1.2 KB)
✔ Applied ServerlessApp to cluster
✔ Provisioning... (this takes ~20-30 seconds)
✔ Ready in 28s

🌐 Endpoint: https://abc123.execute-api.us-east-1.amazonaws.com/prod

Test it:
  curl https://abc123.execute-api.us-east-1.amazonaws.com/prod

Clean up:
  platform destroy hello-world
```

## Test

```bash
curl $(kubectl get serverlessapp hello-world -o jsonpath='{.status.apiEndpoint}')
```

Expected response:

```json
{
  "message": "Hello from PlatFormer!",
  "app": "hello-world",
  "path": "/",
  "timestamp": "2026-03-23T00:00:00.000Z"
}
```

## Check status

```bash
platform status hello-world
```

## View logs

```bash
aws logs tail /aws/lambda/platformer-default-hello-world --follow
```

## Destroy

```bash
platform destroy hello-world
```

PlatFormer deletes the Lambda function, API Gateway, IAM role, and CloudWatch log group via its finalizer.
