---
layout: default
title: Built-in Checks
nav_order: 4
---

# Built-in Checks

This document describes all available built-in checks and their parameters.

**Table of Contents**

- [AWS Checks](#aws-checks)
  - [cloud.aws_authentication](#cloudaws_authentication)
  - [cloud.aws_s3_access](#cloudaws_s3_access)
- [Git Checks](#git-checks)
  - [git.is_up_to_date](#gitis_up_to_date)
- [Kubernetes Checks](#kubernetes-checks)
  - [k8s.namespace_access](#k8snamespace_access)
- [OS Checks](#os-checks)
  - [os.file_exists](#osfile_exists)
  - [os.executable_exists](#osexecutable_exists)

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

## Git Checks

{: #git-checks }

### git.is_up_to_date

Verifies if the current branch contains all changes from the default remote branch. By default, it looks for 'main' or 'master' as the default branch, but you can specify a custom default branch.

**Parameters:**

- `path` (optional): Path to the git repository (defaults to current directory)
- `default_branch` (optional): Name of the default branch to check against (defaults to trying 'main' then 'master')
- `fail_out_of_date` (optional): If true, returns failure status when branch is not up to date. If false or not set, returns warning status.

**Example:**

```yaml
# Basic check using default settings
- name: Check if branch is up to date
  type: git.is_up_to_date

# Check against specific branch and fail if not up to date
- name: Check if branch contains develop changes
  type: git.is_up_to_date
  parameters:
    path: "/path/to/repo"
    default_branch: "develop"
    fail_out_of_date: true
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

### os.executable_exists

Verifies if an executable exists and has proper execution permissions. The check can look for the executable in the system PATH or in a custom directory.

**Parameters:**

- `name` (required): Name of the executable to find
- `custom_path` (optional): Custom directory path to look for the executable. If not provided, only the system PATH is searched.

**Example:**

```yaml
# Check if git is available in PATH
- name: Check git installation
  type: os.executable_exists
  parameters:
    name: git

# Check for executable in custom location
- name: Check custom tool
  type: os.executable_exists
  parameters:
    name: my-tool
    custom_path: /usr/local/bin
```

To author your own checks, see the [Writing Your Own Checks]({% link writing-your-own-checks.md %}) section.
