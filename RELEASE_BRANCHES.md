# Release Branches

Release branches have a name of `v#.#` like `v0.9`. The branch with `v#.#rc` is the release candidate branch. This document describe how a release branch is created and how a release is made. All these steps have to be executed by release manager who has write permission.

## 1. Create release candidate branch

On the feature freeze day, a release candidate branch will be created.

```
git clone https://github.com/opea-project/GenAIInfra.git
cd GenAIInfra
git checkout -b v0.9rc
git push origin v0.9rc
```

## 2. Create images with release tag

This step is being executed under `GenAIExamples`.

In the [Actions](https://github.com/opea-project/GenAIExamples/actions), select the workflow "Examples CD workflow on manual event", and manually trigger this workflow. (Refer to [github website](https://docs.github.com/en/actions/managing-workflow-runs-and-deployments/managing-workflow-runs/manually-running-a-workflow) to get how to manually run a workflow.)

There will be a window promoted for your input. The inputs to the workflow are:
| Description | Value |
|-----------------------------|------------|
|Use workflow from|Branch: v0.9rc|
|Hardware to run test|"gaudi,xeon"|
|List of examples to test|"AudioQnA,ChatQnA,CodeGen,CodeTrans,DocSum,FaqGen,SearchQnA,Translation"|
|Tag to apply to images|"v0.9rc"|
|Whether to deploy gmc|false|
|Build test required images for Examples|true|
|Scan all images with Trivy|false|
|Test examples with docker compose|false|
|Test examples with k8s|false|
|Test examples with gmc|false|
|OPEA branch for image build|"v0.9rc"|

After this workflow executed, the images for GenAIExamples and GenAIComps with `v0.9` tag will be created and pushed to CI local image registries both in Gaudi and Xeon environment.

## 3. Test helm charts

This step is being executed under `GenAIInfra`.

In the [Actions](https://github.com/opea-project/GenAIInfra/actions), select the workflow "GenAIInfra Helm CD workflow on manual event", and manually trigger this workflow. There will be a window promoted for your input.

The inputs to the workflow are:
| Description | Value |
|-----------------------------|------------|
|Use workflow from|Branch: v0.9rc|
|workloads to test, empty for testing all helm charts|""|
|Image tag to be tested|"v0.9"|
|Hardwares used to run tests|"xeon,gaudi"|

All the helm charts will be tested. Green ticks show helm charts pass the tests.

## 4. Test GMC

This step is being executed under `GenAIInfra`. The test will be executed in GMC CI K8s cluster, which is a KIND cluster for now.

In the [Actions](https://github.com/opea-project/GenAIInfra/actions), select the workflow "GenAIInfra GMC CD workflow on manual event", and manually trigger this workflow. There will be a window promoted for your input.

The inputs to the workflow are:
| Description | Value |
|-----------------------------|------------|
|Use workflow from|Branch: v0.9rc|
|Tag to apply to images|"v0.9"|
|Whether to build test required images for GMC|true|

GMC images will be built with v0.9rc branch. GMC test cases will be tested in KIND cluster on Xeon. Green ticks show helm charts pass the tests.

## 5. Publish images

The image pass action will be executed under `GenAIExamples` repo. After all previous tests pass and GenAIExamples release tests also pass, the images will be pushed through GenAIExample workflow [Examples publish docker image on manual event](https://github.com/opea-project/GenAIExamples/actions/workflows/manual-docker-publish.yml).
