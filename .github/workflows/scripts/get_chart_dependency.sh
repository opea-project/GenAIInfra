#!/bin/bash

# Copyright (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

# Function to check dependencies using helm dependency list
check_dependencies() {
    local chart_dir="$1"
    local target_charts=("$@")

    # Get the name of the current chart
    local chart_name
    chart_name=$(basename "$chart_dir")

    # Skip if the current chart is one of the target charts
    for target in "${target_charts[@]:1}"; do
        if [[ "$chart_name" == "$target" ]]; then
            return
        fi
    done

    # Get the list of dependencies using helm dependency list
    local dependencies
    dependencies=$(helm dependency list "$chart_dir" 2>/dev/null | grep 'file://' | awk '{print $1}')

    # Check if any of the dependencies match the target charts
    for dep in $dependencies; do
        for target in "${target_charts[@]:1}"; do
            if [[ "$dep" == "$target" ]]; then
                echo "$chart_dir"
                return
            fi
        done
    done
}

# Main script
if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <chart_directory> [<chart1> <chart2> ... <chartN>]"
    exit 1
fi

chart_directory="$1"
shift
target_charts=("$@")

# Find all Helm charts in the specified directory
chart_dirs=$(find "$chart_directory" -type d -exec test -e {}/Chart.yaml \; -print)

# Iterate over each chart directory and check dependencies
for chart_dir in $chart_dirs; do
    check_dependencies "$chart_dir" "${target_charts[@]}"
done
