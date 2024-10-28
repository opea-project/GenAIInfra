# VideoQnA

Helm chart for deploying VideoQnA service. VideoQnA depends on the following other microservices:

- [data-prep](../common/data-prep/README.md)
- [embedding-usvc](../common/embedding-usvc/README.md)
- [retriever-usvc](../common/retriever-usvc/README.md)
- [reranking-usvc](../common/reranking-usvc/README.md)
- [vdms-vector-db](../common/vdms-vector-db/README.md)
- [lvm-serving](../common/lvm-serving/README.md)
- [lvm-uservice](../common/lvm-uservice/README.md)

## Installing the Chart

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update videoqna

# Set following required values for videoqna chart and various subcharts
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
export CACHEDIR="/home/$USER/.cache"
export LLM_DOWNLOAD=true
export INDEX_NAME="mega-videoqna"

# Set the proxy variables. Assign empty string if no proxy setup required.
export https_proxy="your_http_proxy"
export http_proxy="your_https_proxy"

helm install videoqna videoqna --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} --set global.cacheUseHostPath=${CACHEDIR} --set lvm-serving.llmDownload=${LLM_DOWNLOAD} --set data-prep.indexName=${INDEX_NAME} --set retriever-usvc.indexName=${INDEX_NAME} --set global.https_proxy=${https_proxy} --set global.http_proxy=${http_proxy} --wait
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` and `CACHEDIR` exists on the node where your workload is schedueled. These dirs are used to cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` and `global.cacheUseHostPath` to 'null' if you don't want to cache the models.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Run the command `kubectl port-forward svc/videoqna 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```bash
curl http://localhost:8888/v1/videoqna -H "Content-Type: application/json" -d '{
      "messages": "What is the man doing?",
      "stream": "True"
      }'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service videoqna-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser and head to `http://<k8s-node-ip-address>:${port}` to use VideoQnA via a web UI.

## Values

| Key                                           | Type    | Default                | Description                                                                                                                      |
| --------------------------------------------- | ------- | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| image.repository                              | string  | `"opea/videoqna"`      |                                                                                                                                  |
| service.port                                  | string  | `"8888"`               |                                                                                                                                  |
| global.modelUseHostPath                       | string  | `"/mnt/opea-models"`   | A directory to where model dir for lvm-serving service is mounted.                                                               |
| global.cacheUseHostPath                       | string  | `"/home/$USER/.cache"` | A directory to where cache dir for several services are mounted.                                                                 |
| lvm-serving.llmDownload                       | boolean | `true`                 | This value when true, makes lvm-serving download a model. Change it to false for stopping lvm-serving from re-downloading model. |
| data-prep.indexName, retriever-usvc.indexName | string  | `"mega-videoqna"`      | This value when true, makes lvm-serving download a model. Change it to false for stopping lvm-serving from re-downloading model. |
