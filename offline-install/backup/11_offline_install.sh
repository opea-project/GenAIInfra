#!/bin/bash

# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

registry="registry:5000"

cd ../helm-charts/
helm uninstall chatqna || true

sed -i "s|image: .*nginx:1.27.1|image: $registry/nginx:1.27.1|g" chatqna/templates/nginx-deployment.yaml
sed -i "/  - name: milvus/,+4d" common/data-prep/Chart.yaml
sed -i "/  - name: milvus/,+4d" common/retriever-usvc/Chart.yaml
sed -i "s/^data:/data:\n  HF_HUB_OFFLINE: 'True'/"   common/tgi/templates/configmap.yaml
sed -i "s/^data:/data:\n  HF_HUB_OFFLINE: 'True'/"   common/tei/templates/configmap.yaml
sed -i "s/^data:/data:\n  HF_HUB_OFFLINE: 'True'/"   common/teirerank/templates/configmap.yaml

./update_dependency.sh
helm dependency update chatqna
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"
helm install chatqna chatqna --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} \
    --set image.repository=$registry/opea/chatqna,image.tag=latest \
    --set tgi.image.repository=$registry/huggingface/text-generation-inference,tgi.image.tag=2.4.0-intel-cpu \
    --set tei.image.repository=$registry/huggingface/text-embeddings-inference,tei.image.tag=cpu-1.5 \
    --set teirerank.image.repository=$registry/huggingface/text-embeddings-inference,teirerank.image.tag=cpu-1.5 \
    --set redis-vector-db.image.repository=$registry/redis/redis-stack,redis-vector-db.image.tag=7.2.0-v9 \
    --set retriever-usvc.image.repository=$registry/opea/retriever-redis,retriever-usvc.image.tag=latest \
    --set chatqna-ui.image.repository=$registry/opea/chatqna-ui,chatqna-ui.image.tag=latest,chatqna-ui.containerPort=5173 \
    --set data-prep.image.repository=$registry/opea/dataprep-redis,data-prep.image.tag=latest
