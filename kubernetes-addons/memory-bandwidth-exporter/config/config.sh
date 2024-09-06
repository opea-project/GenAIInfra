#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

sudo apt install jq
sudo mkdir -p /etc/containers/oci/hooks.d/
sudo cp plugin/etc/containers/oci/hooks.d/* /etc/containers/oci/hooks.d/
sudo cp plugin/usr/local/sbin/* /usr/local/sbin/
