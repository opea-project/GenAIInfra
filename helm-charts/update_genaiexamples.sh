#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

CUR_DIR=$(cd $(dirname "$0") && pwd)
MODELPATH="/mnt/opea-models"

GENAIEXAMPLEDIR=${CUR_DIR}/../../GenAIExamples

if [ "f$1" != "f" ]; then
	GENAIEXAMPLEDIR=$1
fi


if [ ! -f $GENAIEXAMPLEDIR/supported_examples.md ]; then
	echo "Can NOT find GenAIExample directory."
	echo "Usage: $0 [GenAIExample dir]"
	exit 1
fi


#
# generate_yaml <chart> <outputdir>
#
function generate_yaml {
  chart=$1
  outputdir=${GENAIEXAMPLEDIR}/$2

  local extraparams=""
  extraparams="--set global.modelUseHostPath=${MODELPATH},image.tag=latest,asr.image.tag=latest,data-prep.image.tag=latest,embedding-usvc.image.tag=latest,llm-uservice.image.tag=latest,reranking-usvc.image.tag=latest,retriever-usvc.image.tag=latest,speecht5.image.tag=latest,tts.image.tag=latest,web-retriever.image.tag=latest,whisper.image.tag=latest"

  helm dependency update $chart
  helm template $chart $chart --skip-tests $extraparams -f $chart/values.yaml       > $outputdir/xeon/${chart}.yaml
  helm template $chart $chart --skip-tests $extraparams -f $chart/gaudi-values.yaml > $outputdir/gaudi/${chart}.yaml

}


${CUR_DIR}/update_dependency.sh
pushd ${CUR_DIR}
generate_yaml chatqna ChatQnA/kubernetes/manifests
generate_yaml codegen CodeGen/kubernetes/manifests
generate_yaml codetrans CodeTrans/kubernetes/manifests
generate_yaml docsum DocSum/kubernetes/manifests
popd
