#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

# Steps for new helm charts release v0.9rc as example.
# This should be implemented in CICD workflow.
# 0. Charts tag updated for this branch
# 1. git clone -b https://github.com/opea-project/GenAIInfra.git
# 2. cd GenAIInfra; git switch v0.9rc; cp -rf helm-charts ../  # Copy the helm-charts directory
# 3. git switch charts-release; mv ../helm-charts . # Switch to charts-release and copy back the helm-charts directory
# 4. ./update_chart_repo.sh; rm -rf helm-charts
# 5. git add .; git commit -s -m "v0.9 charts release"; git push
#  For 1.0 and later, include the gmc helm chart too
# cd microservice-connector/; helm package helm; cp gmc-1.0.0.tgz

UPD_DIR=$(cd $(dirname "$0") && pwd)
CHART_DIR=./helm-charts

# package_chart chartdir
function package_chart {
  CHART=$1
  helm package $1
}

pushd $UPD_DIR
# update components
for chart in $UPD_DIR/$CHART_DIR/common/*
do
	echo "Packaging chart $chart..."
	helm dependency update ${chart}
	helm package $chart
done

# update E2E Applications
for chart in $UPD_DIR/$CHART_DIR/*
do
	if [ -f $chart ]; then continue; fi
	if [[ $chart =~ "common" ]]; then continue; fi
	echo "Packaging chart $chart..."
	helm dependency update ${chart}
	helm package $chart
done

# rm helm-charts
rm -rf helm-charts

# Update Index
# mv index.yaml index.old
helm repo index .

# Insert copyright to avoid warnings
sed -i '1s/^/# Copyright (C) 2024 Intel Corporation\n# SPDX-License-Identifier: Apache-2.0\n\n/' index.yaml
