#!/bin/bash
# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

#set -xe

function check_chart {
  chart=$1
  valuefiles=$(yq eval ".$chart.values" $configfile |sed 's/^- //')

  echo "Checking value files for $chart ..."
  src_repo=$(yq eval ".$chart.src_repo" $configfile)
  dest_repo=$(yq eval ".$chart.dest_repo" $configfile)
  src_dir=$(yq eval ".$chart.src_dir" $configfile)
  dest_dir=$(yq eval ".$chart.dest_dir" $configfile)
  echo $valuefiles
  for valuefile in $valuefiles; do
    echo “  Checking $valuefile”
    # wget https://raw.githubusercontent.com/opea-project/${src_repo}/refs/heads/main/${src_dir}/${valuefile} -qO tmp/src.yaml
    # Use local version src values
    if [[ -d $chart ]]; then
      cp $chart/$valuefile tmp/src.yaml
    else
      cp common/$chart/$valuefile tmp/src.yaml
    fi
    wget https://raw.githubusercontent.com/opea-project/${dest_repo}/refs/heads/main/${dest_dir}/${valuefile} -qO tmp/dest.yaml
    diff tmp/src.yaml tmp/dest.yaml
    rm tmp/src.yaml tmp/dest.yaml
  done
}

configfile=valuefiles.yaml

charts_list=${1:-$(cat valuefiles.yaml |grep -v "^#" |grep -v "^  " |grep -v "^$" |sed 's/:/ /')}

echo $charts_list
mkdir -p tmp
for chart in $charts_list;do
  check_chart $chart
done
