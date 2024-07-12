#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

CUR_DIR=$(cd $(dirname "$0") && pwd)

#
# generate_yaml llm-uservice
#
function generate_yaml {
  chart=$1
  mkdir -p $chart/xeon
  helm template $chart ../helm-charts/common/$chart --values ../helm-charts/common/$chart/values.yaml > $chart/xeon/$chart.yaml
  mkdir -p $chart/gaudi
  if [ -f ../helm-charts/common/$chart/gaudi-values.yaml ]; then
    helm template $chart ../helm-charts/common/$chart --values ../helm-charts/common/$chart/gaudi-values.yaml > $chart/gaudi/$chart.yaml
  else
    helm template $chart ../helm-charts/common/$chart --values ../helm-charts/common/$chart/values.yaml > $chart/gaudi/$chart.yaml
  fi
}

# echo $CUR_DIR
cd $CUR_DIR
for chart in ${CUR_DIR}/../helm-charts/common/*
do
	chartname=`basename $chart`
	echo "Update manifest for $chartname..."
	generate_yaml $chartname
done
