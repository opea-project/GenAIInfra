# ChatQnA

Helm chart for deploying ChatQnA service. ChatQnA depends on the following services:

- [data-prep](../common/data-prep)
- [embedding-usvc](../common/embedding-usvc)
- [tei](../common/tei)
- [retriever-usvc](../common/retriever-usvc)
- [redis-vector-db](../common/redis-vector-db)
- [reranking-usvc](../common/reranking-usvc)
- [teirerank](../common/teirerank)
- [llm-uservice](../common/llm-uservice)
- [tgi](../common/tgi)

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update chatqna
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export MODELNAME="Intel/neural-chat-7b-v3-3"
helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME}
# To use Gaudi device
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/gaudi-values.yaml
# To use Nvidia GPU
#helm install chatqna chatqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set tgi.LLM_MODEL_ID=${MODELNAME} -f chatqna/nv-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Run the command `kubectl port-forward svc/chatqna 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:8888/v1/chatqna \
    -H "Content-Type: application/json" \
    -d '{"messages": "What is the revenue of Nike in 2023?"}'
```

### Verify the workload through UI

UI need to get installed before accessing. Follow the steps below to build and install UI:

```bash
# expose endpoint of ChatQnA service and dataprep service
kubectl port-forward svc/chatqna --address 0.0.0.0 8888:8888
kubectl port-forward svc/chatqna-data-prep --address 0.0.0.0 6007:6007

# build and push the UI image if not exist
# skip these steps if the image already exists
git clone https://github.com/opea-project/GenAIExamples.git
cd GenAIExamples/ChatQnA/docker/ui/
docker build --no-cache -t opea/chatqna-conversation-ui:latest --build-arg https_proxy=$https_proxy --build-arg http_proxy=$http_proxy -f ./docker/Dockerfile.react .
# push the image to your cluster, make sure the image exists on each node of your cluster
docker save -o ui.tar opea/chatqna-conversation-ui:latest
sudo ctr -n k8s.io image import ui.tar

# install UI using helm chart. Replace image tag if required
cd
cd GenAIInfra/helm-charts/
helm install ui common/chatqna-ui --set BACKEND_SERVICE_ENDPOINT="http://${host_ip}:8888/v1/chatqna",DATAPREP_SERVICE_ENDPOINT="http://${host_ip}:6007/v1/dataprep",image.tag="latest"

# expose the endpoint of UI for verification
kubectl port-forward svc/ui --address 0.0.0.0 5174:5174
```

Access `http://localhost:5174` to play with the ChatQnA workload through UI.

## Values

| Key                                    | Type   | Default                       | Description                                                                                                                             |
| -------------------------------------- | ------ | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                       | string | `"opea/chatqna"`              |                                                                                                                                         |
| service.port                           | string | `"8888"`                      |                                                                                                                                         |
| tgi.LLM_MODEL_ID                       | string | `"Intel/neural-chat-7b-v3-3"` | Models id from https://huggingface.co/, or predownloaded model directory                                                                |
| global.horizontalPodAutoscaler.enabled | bop;   | false                         | HPA autoscaling for the TGI and TEI service deployments based on metrics they provide. See HPA section in ../README.md before enabling! |
