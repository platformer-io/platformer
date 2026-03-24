# hello-world

A minimal Node.js Lambda function deployed via PlatFormer. Returns a JSON greeting with a timestamp.

## Prerequisites

- AWS CLI configured (`aws configure`)
- An S3 bucket in the same region as your operator
- A running kind cluster with the PlatFormer CRD installed
- `platform` CLI built (`go build -o bin/platform ./cmd/cli`)

## 1. Zip the function

```bash
cd examples/hello-world/src
zip ../function.zip index.js
```

## 2. Upload to S3

```bash
aws s3 cp examples/hello-world/function.zip \
  s3://YOUR_BUCKET/examples/hello-world/function.zip
```

## 3. Update the manifest

Edit [deploy/serverlessapp.yaml](deploy/serverlessapp.yaml) and replace `REPLACE_WITH_YOUR_BUCKET` with your bucket name:

```yaml
code:
  s3Bucket: my-platformer-bucket   # ← your bucket here
  s3Key: examples/hello-world/function.zip
```

## 4. Deploy

```bash
kubectl apply -f examples/hello-world/deploy/serverlessapp.yaml
```

Or using the PlatFormer CLI:

```bash
platform deploy examples/hello-world/deploy/serverlessapp.yaml
```

## 5. Watch the rollout

```bash
kubectl get serverlessapps -w
```

Once `PHASE` reaches `Ready`, the `ENDPOINT` column shows your live URL.

## 6. Test the endpoint

```bash
curl $(kubectl get serverlessapp hello-world -o jsonpath='{.status.endpoint}')
```

Expected response:

```json
{
  "message": "Hello from PlatFormer!",
  "app": "hello-world",
  "timestamp": "2026-03-23T00:00:00.000Z"
}
```

## 7. Tear down

```bash
kubectl delete serverlessapp hello-world
```

PlatFormer will delete the Lambda function, API Gateway, IAM role, and CloudWatch log group automatically via its finalizer.
