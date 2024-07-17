#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

CUR_DIR=$(cd $(dirname "$0") && pwd)
OUTPUTDIR=${CUR_DIR}/common

#
# generate_yaml <chart> <outputdir>
#
function generate_yaml {
  chart=$1
  outputdir=$2

  helm template $chart ../helm-charts/common/$chart --skip-tests --values ../helm-charts/common/$chart/values.yaml --set global.extraEnvConfig=extra-env-config,noProbe=true > ${outputdir}/$chart.yaml
  if [ -f ../helm-charts/common/$chart/gaudi-values.yaml ]; then
    helm template $chart ../helm-charts/common/$chart --skip-tests --values ../helm-charts/common/$chart/gaudi-values.yaml --set global.extraEnvConfig=extra-env-config,noProbe=true  > ${outputdir}/${chart}_gaudi.yaml
  fi
}

mkdir -p $OUTPUTDIR
${CUR_DIR}/../helm-charts/update_dependency.sh
cd $CUR_DIR
for chart in ${CUR_DIR}/../helm-charts/common/*
do
	chartname=`basename $chart`
	echo "Update manifest for $chartname..."
	generate_yaml $chartname $OUTPUTDIR
done

# we need special version of docsum-llm-uservice
echo "Update manifest for docsum-llm-uservice..."
helm template docsum ../helm-charts/common/llm-uservice --skip-tests --set global.extraEnvConfig=extra-env-config,noProbe=true,image.repository=opea/llm-docsum-tgi:latest > ${OUTPUTDIR}/docsum-llm-uservice.yaml
