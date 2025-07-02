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

# needs to be specified so there's something to override
NS="default"

error_exit ()
{
	name=${0##*/}
	cat << EOF

ERROR: $1!

Convert given Grafana *.json dashboard specs to Grafana dashboard configMap
templates compatible with OPEA Dashoard Helm chart.

Dashboard 'title' and 'uid' are overridden with values composed of
the OPEA Dashboard Helm chart values and file name, similarly to
the produced configMap name and namespace.

Usage:
	$name <JSON files>

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

uid=""
title=""

echo
echo "Got following Grafana dashboards:"
for file in "$@"; do
	if [ ! -f "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi
	if [ "${file%.json}" = "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi

	# Both dashboard 'uid' and title needed
	uid=$(jq .uid "$file" | tail -1 | tr -d '"')
	if [ -z "$uid" ]; then
		error_exit "'$file' dashboard has invalid JSON"
	elif [ "$uid" = "null" ]; then
		error_exit "'$file' dashboard has no 'uid' field (will be replaced with Helm variable)"
	fi

	# ...but it should have a title.
	title=$(jq .title "$file" | tail -1 | tr -d '"')
	if [ "$title" = "null" ]; then
		error_exit "'$file' dashboard has no 'title' field (will be replaced with Helm variable)"
	fi

	echo "- file: $file, uid: '$uid', title: '$title'"
done

echo
echo "Converting:"
for file in "$@"; do
	base=${file##*/}
	name=${base%.json}
	dst="configmap-${name}.yaml"

	uid=$(jq .uid "$file" | tail -1 | tr -d '"')
	title=$(jq .title "$file" | tail -1 | tr -d '"')

	# convert to k8s object name ("[a-z0-9][-a-z0-9]*[a-z0-9]"):
	# - upper-case -> lowercase, '_' -> '-'
	# - drop anything outside [-a-z]
	# - drop '-' prefix & suffix and successive '-' chars
	k8name=$(echo "$name" | tr A-Z_ a-z- | tr -d -c a-z- | sed -e 's/^-*//' -e 's/-*$//' -e 's/--*/-/g')

	echo "- $base -> $dst"

	echo "{{- if .Values.$name }}" > "$dst"

	echo "# $COPYRIGHT" >> "$dst"
	echo "# $LICENSE" >> "$dst"
	echo "#" >> "$dst"
	echo "# ${0##*/}: $base -> $dst" >> "$dst"

	kubectl create cm -n "$NS" --from-file "$file" --dry-run=client -o yaml "$k8name" |\
	  kubectl label -f- --local --dry-run=client -o yaml "$LABEL" |\
	  grep -v -e "^  creationTimestamp:" >> "$dst"

	echo "{{- end }}" >> "$dst"

	# convert JSON content conflicting with Helm to Helm compatible format
	# and add suitable Dashboard chart Helm variables to the configMap
	sed -i \
	  -e 's/\({{[a-z]\+}}\)/{{ printf "\1" }}/' \
	  -e 's/name:.*$/name: {{ include "dashboard.fullname" . }}'"-${k8name}/" \
	  -e 's/space:.*$/space: {{ .Values.global.prometheusNamespace }}/' \
	  -e "s/${title}/{{ .Values.prefix }} $name/" \
	  -e "s/${uid}/opea-"'{{ include "dashboard.fullname" . }}'"-${k8name}/" \
	  "$dst"
done

echo
echo "DONE!"
