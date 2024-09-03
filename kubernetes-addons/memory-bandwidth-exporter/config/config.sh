#!/bin/bash

sudo apt install jq
sudo mkdir -p /etc/containers/oci/hooks.d/
sudo cp plugin/etc/containers/oci/hooks.d/* /etc/containers/oci/hooks.d/
sudo cp plugin/usr/local/sbin/* /usr/local/sbin/