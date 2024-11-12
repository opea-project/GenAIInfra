#!/bin/bash


registry=registry:5000

# Define the images to pull and push
images=(
  "docker.io/nginx:1.27.1"
  "docker.io/opea/chatqna-conversation-ui:latest"
  "docker.io/opea/chatqna:latest"
  "docker.io/opea/chatqna-ui:latest"
  "docker.io/opea/dataprep-redis:latest"
  "docker.io/opea/retriever-redis:latest"
  "docker.io/redis/redis-stack:7.2.0-v9"
  "gcr.io/google_containers/pause:3.1"
  "ghcr.io/huggingface/text-embeddings-inference:cpu-1.5"
  "ghcr.io/huggingface/text-generation-inference:2.4.0-intel-cpu"
)

# Loop through each image, pull it, tag it, and push it to the registry
for image in "${images[@]}"; do
  # Extract the image name and tag
  image_name=$(echo "$image" | awk -F ':' '{print $1}' | awk -F '/' ' { for (i=2; i<=NF-1; i++) printf $i "/";  print $(NF) }')
  image_tag=$(echo "$image" | awk -F: '{print $2}')
  
  # If no tag is specified, default to latest
  if [ -z "$image_tag" ]; then
    image_tag="latest"
  fi

  echo "Pulling image: $image"
  docker pull "$image"
  
  # Tag the image for the new registry with the correct tag
  new_image="$registry/$image_name:$image_tag"
  echo "Tagging image: $image as $new_image"
  docker tag "$image" "$new_image"
  
  # Push the image to the new registry
  echo "Pushing image: $new_image"
  docker push "$new_image"
done
