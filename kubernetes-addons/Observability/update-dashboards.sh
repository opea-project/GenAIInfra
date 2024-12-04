#!/bin/sh
#
# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

set -e

# Grafana namespace
ns="monitoring"

# Grafana app selector
selector="app.kubernetes.io/name=grafana"

# command for fetching Grafana pod name
grafana="kubectl -n $ns get pod --selector $selector --field-selector=status.phase=Running -o name"

# Labels needed in configMap to get (Helm installed) Grafana to load it as dashboard
labels="grafana_dashboard=1"

error_exit ()
{
	name=${0##*/}
	echo
	echo "Create/update dashboards specified in given JSON files for Grafana."
	echo "Names for the created configMaps will be in '$USER-<filename>' format"
	echo "and they go to (Grafana) '$ns' namespace."
	echo
	echo "usage: $name <JSON files>"
	echo
	echo "ERROR: $1!"
	exit 1
}

if [ $# -lt 1 ]; then
	error_exit "no files specified"
fi

if [ -z "$(which jq)" ]; then
	echo "ERROR: 'jq' required for dashboard checks, please install it first!"
	exit 1
fi

echo "Creating/updating following Grafana dashboards to '$ns' namespace:"
for file in "$@"; do
	if [ ! -f "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi
	# Dashboard 'uid' is optional as Grafana can generate one...
	uid=$(jq .uid "$file" | tail -1)
	if [ -z "$uid" ]; then
		error_exit "'$file' dashboard has invalid JSON"
	fi
	# ...but it should have a title.
	title=$(jq .title "$file" | tail -1)
	if [ "$title" = "null" ]; then
		error_exit "'$file' dashboard has no 'title' field"
	fi
	echo "- $file (uid: $uid): $title"
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

pod=$($grafana)
if [ -z "$pod" ]; then
	echo "ERROR: Grafana missing from '$ns' namespace!"
	exit
fi

echo
for file in "$@"; do
	base=${file##*/}
	name=${base%.json}
	# if no user prefix, add one
	if [ "${name#"$USER"}" = "$name" ]; then
		name="$USER-$name"
	fi
	# convert to k8s object name ("[a-z0-9][-a-z0-9]*[a-z0-9]"):
	# - upper-case -> lowercase, '_' -> '-'
	# - drop anything outside [-a-z]
	# - drop '-' prefix & suffix and successive '-' chars
	name=$(echo "$name" | tr A-Z_ a-z- | tr -d -c a-z- | sed -e 's/^-*//' -e 's/-*$//' -e 's/--*/-/g')
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

echo "DONE! => Dashboard(s) should appear in Grafana after 1 min wait *and* Grafana page reload."
