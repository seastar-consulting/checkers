---
layout: default
title: Built-in Checks
nav_order: 3
---

# Built-in Checks

This document describes all available built-in checks and their parameters.

**Table of Contents**
- [AWS Checks](#aws-checks)
  - [cloud.aws_authentication](#cloudaws_authentication)
  - [cloud.aws_s3_access](#cloudaws_s3_access)
- [Kubernetes Checks](#kubernetes-checks)
  - [k8s.namespace_access](#k8snamespace_access)
- [OS Checks](#os-checks)
  - [os.file_exists](#osfile_exists)

## AWS Checks
{: #aws-checks }

### cloud.aws_authentication

Verifies AWS credentials and identity by calling the STS GetCallerIdentity API.

**Parameters:**
- `aws_profile` (optional): AWS profile to use
- `identity` (required): Expected AWS ARN to match against

**Example:**
```yaml
- name: verify-aws-identity
  type: cloud.aws_authentication
  parameters:
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
  parameters:
    bucket: "my-bucket"
    aws_profile: "prod"

# Check read access to specific object
- name: check-s3-object-read
  type: cloud.aws_s3_access
  parameters:
    bucket: "my-bucket"
    key: "path/to/file.txt"
    aws_profile: "prod"
```

## Kubernetes Checks
{: #kubernetes-checks }

### k8s.namespace_access

Verifies access to a Kubernetes namespace by attempting to list pods in that namespace.

**Parameters:**
- `namespace` (optional): Kubernetes namespace to check (defaults to "default")
- `context` (optional): Kubernetes context to use

**Example:**
```yaml
- name: verify-k8s-access
  type: k8s.namespace_access
  parameters:
    namespace: "production"
    context: "prod-cluster"
```

## OS Checks
{: #os-checks }

### os.file_exists

Verifies if a file exists at the specified path.

**Parameters:**
- `path` (required): The file path to check

**Example:**
```yaml
- name: check-config-file
  type: os.file_exists
  parameters:
    path: "/path/to/config.yaml"
```

To author your own checks, see the [Writing Your Own Checks]({% link writing-your-own-checks.md %}) section.