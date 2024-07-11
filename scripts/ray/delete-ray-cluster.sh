#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -exo pipefail

kubectl delete -f $DIR/ray-cluster-autoscaler.yaml
