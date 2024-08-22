#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

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

# Update Index
helm repo index .
