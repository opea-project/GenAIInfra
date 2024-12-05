#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

CUR_DIR=$(cd $(dirname "$0") && pwd)
OUTPUTDIR=${CUR_DIR}/../microservices-connector/config/manifests
MODELPATH="/mnt/opea-models"

NEWTAG="${NEWTAG:-latest}"
#
# generate_yaml <chart> <outputdir>
#
function generate_yaml {
  chart=$1
  outputdir=$2

  local extraparams=""
  if [[ $(grep -c 'tag: "latest"' ./common/$chart/values.yaml) != 0 ]]; then
     extraparams="--set image.tag=$NEWTAG"
  fi

  helm template $chart ./common/$chart --skip-tests --values ./common/$chart/values.yaml --set global.extraEnvConfig=extra-env-config,global.modelUseHostPath=$MODELPATH $extraparams > ${outputdir}/$chart.yaml

  for f in `ls ./common/$chart/*-values.yaml 2>/dev/null `; do
    filename=$(basename $f)
    releasename=$chart
    if [[ $filename == ci-* ]]; then
      continue
    fi
    if [[ "$filename" =~ ^variant_.*-values.yaml ]]; then
      ext=$(echo $filename | sed 's/^variant_//' | sed 's/-values.yaml$//')
      outputfile="$ext-${chart}.yaml"
      releasename=$(echo "${ext}-${chart}" | sed 's/_/-/g')
    else
      ext=$(echo $filename | sed 's/-values.yaml$//')
      outputfile="${chart}_${ext}.yaml"
    fi
    extraparams=""
    if [[ $(grep -c 'tag: "latest"' $f) != 0 ]]; then
       extraparams="--set image.tag=$NEWTAG"
    fi
    helm template $releasename ./common/$chart --skip-tests --values ${f} --set global.extraEnvConfig=extra-env-config,global.modelUseHostPath=$MODELPATH $extraparams > ${outputdir}/${outputfile}
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
