---
layout: home
title: Home
nav_order: 1
---

# Checkers

Checkers is a diagnostics framework for developer workstations. It helps ensure
that your development environment is correctly configured and running smoothly.

## Quick Links

- [Available Check Types]({% link check-types.md %}): Learn about all built-in checks and their parameters
- [Writing Custom Checks]({% link writing-your-own-checks.md %}): Learn how to create and integrate your own checks

## Installation

You can install Checkers in one of two ways:

1. Using Go:
```bash
go install github.com/seastar-consulting/checkers@latest
```

2. Download the binary from [GitHub]({{ site.aux_links['Checkers on GitHub'][0] }}/releases/latest)
and add it to your PATH.

## Usage

First, you need to create a configuration file named `checkers.yaml` in your
current directory. This file should contain the checks to be run and their
configuration.

Here is an example of a `checks.yaml` file:

```yaml
checks:
  - name: verify-aws-identity
    type: cloud.aws_authentication
    params:
      aws_profile: "prod"
      identity: "arn:aws:iam::123456789012:user/myuser"

  - name: check-s3-bucket
    type: cloud.aws_s3_access
    params:
      bucket: "my-bucket"

  - name: verify-k8s-access
    type: k8s.namespace_access
    params:
      namespace: "production"
      context: "prod-cluster"
```

You can run Checkers using the following command:

> checkers

Checkers will run a series of checks on your development environment and provide
a summary of the results.

For more detailed information about available checks and configuration options,
check out our [Getting Started Guide]({% link getting-started.md %}).
