#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

CUR_DIR=$(cd $(dirname "$0") && pwd)
OUTPUTDIR=${CUR_DIR}/../microservices-connector/config/manifests
MODELPATH="/mnt/opea-models"

#
# generate_yaml <chart> <outputdir>
#
function generate_yaml {
  chart=$1
  outputdir=$2

  local extraparams=""
  if [[ $(grep -c 'tag: ""' ./common/$chart/values.yaml) != 0 ]]; then
     extraparams="--set image.tag=latest"
  fi

  helm template $chart ./common/$chart --skip-tests --values ./common/$chart/values.yaml --set global.extraEnvConfig=extra-env-config,global.modelUseHostPath=$MODELPATH,noProbe=true $extraparams > ${outputdir}/$chart.yaml

  for f in `ls ./common/$chart/*-values.yaml 2>/dev/null `; do
    ext=$(basename $f | cut -d'-' -f1)
    extraparams=""
    if [[ $(grep -c 'tag: ""' $f) != 0 ]]; then
       extraparams="--set image.tag=latest"
    fi
    helm template $chart ./common/$chart --skip-tests --values ${f} --set global.extraEnvConfig=extra-env-config,global.modelUseHostPath=$MODELPATH,noProbe=true $extraparams > ${outputdir}/${chart}_${ext}.yaml
  done
}

mkdir -p $OUTPUTDIR
${CUR_DIR}/update_dependency.sh
cd $CUR_DIR
for chart in ${CUR_DIR}/common/*
do
	chartname=`basename $chart`
	echo "Update manifest for $chartname..."
	generate_yaml $chartname $OUTPUTDIR
done

# we need special version of docsum-llm-uservice
echo "Update manifest for docsum-llm-uservice..."
helm template docsum ./common/llm-uservice --skip-tests --set global.extraEnvConfig=extra-env-config,global.modelUseHostPath=$MODELPATH,noProbe=true,image.repository=opea/llm-docsum-tgi,image.tag=latest> ${OUTPUTDIR}/docsum-llm-uservice.yaml
