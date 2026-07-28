# hello-world

A multi-route Node.js web application deployed via PlatFormer. Serves a styled
HTML dashboard at `/`, a JSON health endpoint at `/health`, and full runtime
info at `/api/info`. Demonstrates real HTTP routing, environment variable
injection, and browser-viewable output — not just a response code.

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
✔ Uploaded function code (3.1 KB)
✔ Applied ServerlessApp to cluster
✔ Provisioning... (this takes ~20-30 seconds)
✔ Ready in 28s

🌐 Endpoint: https://abc123.execute-api.us-east-1.amazonaws.com/prod

Test it:
  curl https://abc123.execute-api.us-east-1.amazonaws.com/prod

Clean up:
  platform destroy hello-world
```

## Endpoints

| Route | Type | What it tests |
|---|---|---|
| `GET /` | HTML dashboard | Routing, HTML responses, env var injection, Lambda context |
| `GET /health` | JSON | Health check path, uptime tracking |
| `GET /api/info` | JSON | Full runtime + request context, structured output |
| `GET /anything-else` | JSON 404 | Error path handling |

## Test

Set the endpoint in a variable for convenience:

```bash
ENDPOINT=$(kubectl get serverlessapp hello-world -o jsonpath='{.status.apiEndpoint}')
```

**Homepage (open in browser or curl):**
```bash
curl $ENDPOINT/
# → styled HTML page with runtime dashboard
```

**Health check:**
```bash
curl $ENDPOINT/health
```
```json
{
  "status": "ok",
  "uptimeMs": 142,
  "timestamp": "2026-05-20T12:00:00.000Z"
}
```

**Runtime info:**
```bash
curl $ENDPOINT/api/info
```
```json
{
  "function": {
    "name": "platformer-default-hello-world",
    "region": "us-east-1",
    "memoryMB": 128
  },
  "app": {
    "env": "production",
    "version": "1.0.0"
  },
  "request": {
    "path": "/api/info",
    "method": "GET",
    "userAgent": "curl/8.4.0"
  },
  "timestamp": "2026-05-20T12:00:00.000Z"
}
```

**404 path:**
```bash
curl -i $ENDPOINT/not-a-route
# → HTTP 404  {"error":"Not found","path":"/not-a-route"}
```

## Check status

```bash
platform status hello-world
```

## View logs

```bash
aws logs tail /aws/lambda/platformer-default-hello-world --follow
```

Logs show every route hit:
```
GET /
GET /health
GET /api/info
```

## Destroy

```bash
platform destroy hello-world
```

PlatFormer deletes the Lambda function, API Gateway, IAM role, and CloudWatch
log group via its finalizer.
