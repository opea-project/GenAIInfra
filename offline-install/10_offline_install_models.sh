#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

registry="registry:5000"

sed -i "s|image:.*models.*|image: $registry/opea/models:latest|" models.yaml
sed -i "s|image:.*pause.*|image: $registry/google_containers/pause:3.1|" models.yaml
kubectl delete -f models.yaml || true
kubectl apply -f models.yaml
