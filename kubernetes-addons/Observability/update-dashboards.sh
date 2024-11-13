#!/bin/sh
#
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# Grafana namespace
ns=monitoring

# Labels needed in configMap to get (Helm installed) Grafana to load it as dashboard
labels="grafana_dashboard=1 release=prometheus-stack app=kube-prometheus-stack-grafana"

usage ()
{
	name=${0##*/}
	echo
	echo "Create/update dashboards specified in given JSON files to Grafana."
	echo "Names for the created configMaps will be in '$USER-<filename>' format."
	echo
	echo "usage: $name <JSON files>"
	echo
	echo "ERROR: $1!"
	exit 1
}

if [ $# -lt 1 ]; then
	usage "no files specified"
fi

for file in "$@"; do
	if [ ! -f "$file" ]; then
		usage "JSON file '$file' does not exist"
	fi
done

if [ -z "$(which jq)" ]; then
	echo "ERROR: 'jq' required for dashboard checks, please install it first!"
	exit 1
fi

echo "Creating/updating following Grafana dashboards to '$ns' namespace:"
for file in "$@"; do
	echo "- $file ($(jq .uid "$file" | tail -1)): $(jq .title "$file" | tail -1)"
done

# use tmp file so user can check what's wrong when there are errors
sep="--------------------------------------"
tmp=_dashboard-tmp.yaml
cleanup ()
{
	set +x
	if [ -f $tmp ]; then
		echo $sep
		cat $tmp
		echo $sep
		rm $tmp
		echo "ERROR: fail, probably in uploading above dashboard"
	fi
}
trap cleanup EXIT

echo
for file in "$@"; do
	base=${file##*/}
	name=${base%.json}
	name="$USER-$name"
	echo "*** $ns/$name: $(jq .title "$file" | tail -1) ***"
	set -x
	# shellcheck disable=SC2086
	kubectl create cm -n "$ns" --from-file "$file" --dry-run=client -o yaml "$name" |\
	  kubectl label -f- --local --dry-run=client -o yaml $labels > $tmp
	kubectl create -f $tmp || kubectl replace -f $tmp
	set +x
	echo
done

rm $tmp

echo "DONE!"
