#!/bin/bash
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

helm uninstall kuberay-operator
kubectl get pods

helm repo remove kuberay
