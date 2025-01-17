# Check Types

This document describes all available check types and their parameters.

## AWS Checks

### cloud.aws_authentication

Verifies AWS credentials and identity by calling the STS GetCallerIdentity API.

**Parameters:**
- `aws_profile` (optional): AWS profile to use
- `identity` (required): Expected AWS ARN to match against

**Example:**
```yaml
- name: verify-aws-identity
  type: cloud.aws_authentication
  params:
    aws_profile: "prod"
    identity: "arn:aws:iam::123456789012:user/myuser"
```

### cloud.aws_s3_access

Verifies access to an S3 bucket. If a key is provided, it verifies read access to that specific object. Otherwise, it creates a test object, verifies write access, and then cleans up.

**Parameters:**
- `bucket` (required): S3 bucket name
- `key` (optional): Specific object to check for read access
- `aws_profile` (optional): AWS profile to use

**Example:**
```yaml
# Check write access
- name: check-s3-bucket-write
  type: cloud.aws_s3_access
  params:
    bucket: "my-bucket"
    aws_profile: "prod"

# Check read access to specific object
- name: check-s3-object-read
  type: cloud.aws_s3_access
  params:
    bucket: "my-bucket"
    key: "path/to/file.txt"
    aws_profile: "prod"
```

## Kubernetes Checks

### k8s.namespace_access

Verifies access to a Kubernetes namespace by attempting to list pods in that namespace.

**Parameters:**
- `namespace` (optional): Kubernetes namespace to check (defaults to "default")
- `context` (optional): Kubernetes context to use

**Example:**
```yaml
- name: verify-k8s-access
  type: k8s.namespace_access
  params:
    namespace: "production"
    context: "prod-cluster"
```

## Check Results

All checks return results in a standard format:

```go
map[string]interface{}{
    "status": "Success" | "Failure",
    "output": "Human readable message",
    // Optional additional fields specific to the check
}
```

The CLI will format these results according to the output mode:

### Default Output
```
✅ verify-aws-identity
✅ check-s3-bucket
❌ verify-k8s-access
```

### Verbose Output (-v flag)
```
✅ verify-aws-identity
   └── Successfully authenticated with AWS

✅ check-s3-bucket
   └── Successfully verified write access to bucket my-bucket

❌ verify-k8s-access
   └── Error accessing namespace production: forbidden
```
