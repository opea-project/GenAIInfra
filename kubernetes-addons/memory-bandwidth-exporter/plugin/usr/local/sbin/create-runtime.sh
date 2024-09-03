#!/bin/bash

container_state=$(cat)
container_pid=$(echo "$container_state" | jq -r '.pid')

container_id=$(echo "$container_state" | jq -r '.id')
mon_group_dir="/sys/fs/resctrl/mon_groups/container-$container_id"
if [ ! -d "$mon_group_dir" ]; then
    mkdir -p "$mon_group_dir"
fi

tasks_file="$mon_group_dir/tasks"
echo "$container_pid" | sudo tee -a $tasks_file > /dev/null
echo "CREATE Container: $container_id, Container State: $container_state, Container PID: $container_pid" | sudo tee -a /var/log/container_pids.log > /dev/null