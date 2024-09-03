#!/bin/bash

container_state=$(cat)
container_id=$(echo "$container_state" | jq -r '.id')
mon_group_dir="/sys/fs/resctrl/mon_groups/container-$container_id"
echo "DELETE Container: $container_id" >> /var/log/container_pids.log
sudo rmdir $mon_group_dir