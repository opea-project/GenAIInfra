# Development

This doc explains how to set up a development environment, so you can get started
[contributing](https://github.com/opea-project/docs/tree/main/community/CONTRIBUTING.md) to `GenAIInfra` and how to test your changes before
submitting a pull request.

## Prerequisites

You must install these tools:

1. [`helm`](https://helm.sh/docs/intro/install/): the package manager for Kubernetes
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/): For
   managing development environments.
1. [`pre-commit`](https://pre-commit.com/#install): better to do a pre-commit test before
   you create a commit

## Testing

To test your changes after developing or updating a helm chart, you can use the following command:

```
make test
```

## pre-commit testing

To run pre-commit tests before creating a commit, execute the following command:

```
make pre-commit
```
