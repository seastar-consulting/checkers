---
layout: home
title: Home
nav_order: 1
---

# Checkers

Checkers is a diagnostics framework for developer workstations. It helps ensure
that your development environment is correctly configured and running smoothly.

## Installation

You can get the latest version from [GitHub]({{ site.aux_links['Checkers on GitHub'][0] }}/releases/latest)
and copy the binary to your PATH.

## Usage

First, you need to create a configuration file named `checkers.yaml` in your
current directory. This file should contain the checks to be run and their
configuration.

Here is an example of a `checks.yaml` file:

```yaml
checks:
    - name: Shell check
      type: command
      command: echo '{"status":"success","output":"test output"}'

    - name: check that this file exists!
      type: os.file_exists
      parameters:
        path: checks.yaml
```

You can run Checkers using the following command:

> checkers

Checkers will run a series of checks on your development environment and provide
a summary of the results.

For more detailed information about available checks and configuration options,
check out our [Getting Started Guide]({% link getting-started.md %}).
