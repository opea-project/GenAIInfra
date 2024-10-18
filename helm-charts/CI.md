# CI guidelines for helm charts

## Table of Contents

- [Infra Setup](#infra-setup)
- [Add new test case](#add-new-test-case)

## Infra Setup

Helm charts are organized as:
1). E2E examples are directly located under helm-charts/ directory
2). Charts for components are located under helm-charts/common directory.

We have Xeon servers with Gaudi accelerator card, which covers CPU and Gaudi accelerated workloads. For coverage of other configurations/features, it's case by case.

CI tests values files which names are listed in `ci-*values.yaml` files.

## Add new test case

Add a new values file following the naming rules: `ci-*values.yaml`.
If the values file for CI is identical to another one, a symbolic link can be used.

See following files for the test scripts:
https://github.com/opea-project/GenAIInfra/blob/main/.github/workflows/pr-chart-e2e.yaml
https://github.com/opea-project/GenAIInfra/blob/main/.github/workflows/_helm-e2e.yaml
