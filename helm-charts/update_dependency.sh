#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

UPD_DIR=$(cd $(dirname "$0") && pwd)
for chart in ${UPD_DIR}/common/*
do
	echo "Update dependency for `basename $chart`..."
        rm -f ${chart}/Chart.lock
        rm -rf ${chart}/charts/
	helm dependency update ${chart}
done
