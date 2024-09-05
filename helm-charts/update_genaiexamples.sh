#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

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
# generate_yaml <chart> <valuefile> <GenAIExample outputdir> <output file name> <extra param passed to --set>
#
function generate_yaml {
  chart=$1
  valuefile=$2
  outputdir=${GENAIEXAMPLEDIR}/$3
  outputfile=$4
  extra=$5

  local extraparams=""
  [[ "x$extra" != "x" ]] && extraparams="--set ${extra}"

  helm dependency update $chart
  helm template $chart $chart --skip-tests $extraparams -f $chart/${valuefile} > $outputdir/$outputfile

}


${CUR_DIR}/update_dependency.sh
pushd ${CUR_DIR}
generate_yaml chatqna 	values.yaml	 		ChatQnA/kubernetes/manifests/xeon	chatqna.yaml
generate_yaml chatqna 	guardrails-values.yaml	 	ChatQnA/kubernetes/manifests/xeon	chatqna-guardrails.yaml
generate_yaml chatqna 	gaudi-values.yaml 		ChatQnA/kubernetes/manifests/gaudi	chatqna.yaml
generate_yaml chatqna 	guardrails-gaudi-values.yaml 	ChatQnA/kubernetes/manifests/gaudi	chatqna-guardrails.yaml
generate_yaml codegen 	values.yaml			CodeGen/kubernetes/manifests/xeon	codegen.yaml
generate_yaml codegen 	gaudi-values.yaml		CodeGen/kubernetes/manifests/gaudi	codegen.yaml
generate_yaml codetrans values.yaml			CodeTrans/kubernetes/manifests/xeon	codetrans.yaml
generate_yaml codetrans gaudi-values.yaml		CodeTrans/kubernetes/manifests/gaudi	codetrans.yaml
generate_yaml docsum	values.yaml			DocSum/kubernetes/manifests/xeon	docsum.yaml
generate_yaml docsum	gaudi-values.yaml		DocSum/kubernetes/manifests/gaudi	docsum.yaml
popd
