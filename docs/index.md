---
layout: default
title: Home
nav_order: 1
---

# Checkers

Checkers is a diagnostics framework. It helps ensure a set of criteria is met
through declarative checklists.

It comes with a simple command-line interface that allows you to run a set of
checks on your system and display the results in a human-readable format. It
includes a variety of built-in checks for common tasks, such as checking for the
presence of required files, verifying access to AWS S3, and more.

Checkers generates reports that summarize results and enables developers to
share their results with their team when they encounter issues.  This
drastically simplifies the debugging process and clearly identifies what needs
to be addressed.

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
  # Built-in checks
  - name: Check if .env file exists in current directory
    type: "os.file_exists"
    parameters:
      path: ".env"

  - name: check-s3-bucket
    type: cloud.aws_s3_access
    parameters:
      bucket: "my-bucket"

  - name: verify-k8s-access
    type: k8s.namespace_access
    parameters:
      namespace: "production"
      context: "prod-cluster"

  # Custom shell checks
  - name: "Check Docker CLI Installation"
    type: "command"
    command: |
      if command -v docker >/dev/null 2>&1; then
        echo '{"status": "success", "output": "Docker CLI is installed"}'
      else
        echo '{"status": "failure", "output": "Docker CLI is not installed"}'
      fi
```

You can run Checkers using the following command:

> checkers

Checkers will run a series of checks on your development environment and provide
a summary of the results.

For more detailed information about available checks and configuration options,
check out our [Getting Started Guide]({% link getting-started.md %}).

## Quick Links

- [Configuration]({% link configuration.md %}): Learn about the configuration options available for Checkers
- [Built-in Checks]({% link built-in-checks.md %}): Learn about all built-in checks and their parameters
- [Writing Custom Checks]({% link writing-your-own-checks.md %}): Learn how to create and integrate your own checks
