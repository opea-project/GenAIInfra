#!/bin/sh
#
# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# OPEA CI requires copyright/license for the converted file
COPYRIGHT="Copyright (C) 2025 Intel Corporation"
LICENSE="SPDX-License-Identifier: Apache-2.0"

# Label needed in configMap to get (Helm installed) Grafana to load it as dashboard
LABEL="grafana_dashboard=1"

PREFIX="opea-"
NS=""

error_exit ()
{
	name=${0##*/}
	cat << EOF

ERROR: $1!

Convert given Grafana *.json dashboard specs to Grafana (Helm install) *.yaml configMaps.

Usage:
	$name <JSON files>

Example:
	$name vllm.json
	kubectl apply -n monitoring vllm.yaml

=> Creates 'vllm.yaml' configMap named as '${PREFIX}vllm' for Grafana in 'monitoring' namespace.

ERROR: $1!
EOF
	exit 1
}

if [ -z "$(which jq)" ]; then
	error_exit "'jq' required for dashboard checks, please install 'jq' first"
fi

if [ -z "$(which kubectl)" ]; then
	error_exit "'kubectl' required for dashboard conversion, please install 'kubernetes-client' first"
fi

if ! kubectl version; then
	error_exit "Broken/missing 'kubectl' cluster config (script does not need it, but kubectl still fails)"
fi

echo
echo "Got following Grafana dashboards:"
for file in "$@"; do
	if [ ! -f "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi
	if [ "${file%.json}" = "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi

	# Dashboard 'uid' is optional as Grafana can generate one...
	uid=$(jq .uid "$file" | tail -1 | tr -d '"')
	if [ -z "$uid" ]; then
		error_exit "'$file' dashboard has invalid JSON"
	elif [ "$uid" = "null" ]; then
		echo "WARNING: no dashboard 'uid', Grafana will assign new one on every load: $file"
	elif echo "$uid" | grep -q -v '^[-0-9a-f]*$'; then
		echo "DEBUG: dashboard 'uid' not in hex format: '$uid'?"
	fi

	# ...but it should have a title.
	title=$(jq .title "$file" | tail -1 | tr -d '"')
	if [ "$title" = "null" ]; then
		error_exit "'$file' dashboard has no 'title' field"
	fi

	echo "- file: $file, uid: '$uid', title: '$title'"
done

echo
echo "Converting:"
for file in "$@"; do
	base=${file##*/}
	name=${base%.json}
	dst="${name}.yaml"

	# if no prefix, add one
	if [ "${name#"$PREFIX"}" = "$name" ]; then
		name="${PREFIX}${name}"
	fi

	# convert to k8s object name ("[a-z0-9][-a-z0-9]*[a-z0-9]"):
	# - upper-case -> lowercase, '_' -> '-'
	# - drop anything outside [-a-z]
	# - drop '-' prefix & suffix and successive '-' chars
	k8name=$(echo "$name" | tr A-Z_ a-z- | tr -d -c a-z- | sed -e 's/^-*//' -e 's/-*$//' -e 's/--*/-/g')

	echo "- file: $dst, configMap: $k8name, title: $(jq .title "$file" | tail -1)"

	echo "# $COPYRIGHT" > "$dst"
	echo "# $LICENSE" >> "$dst"
	echo "#" >> "$dst"
	echo "# ${0##*/}: $base -> $dst" >> "$dst"

	kubectl create cm -n "$NS" --from-file "$file" --dry-run=client -o yaml "$k8name" |\
	  kubectl label -f- --local --dry-run=client -o yaml "$LABEL" |\
	  grep -v -e "^  creationTimestamp:" >> "$dst"
done

echo
echo "DONE!"
