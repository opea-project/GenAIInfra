# FaqGen

Helm chart for deploying FaqGen service.

FaqGen depends on LLM and tgi microservice, refer to [llm-uservice](../common/llm-uservice/README.md) and [tgi](../common/tgi/README.md) for more config details.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/faqgen 8888:8888` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:8888/v1/faqgen \
      -H "Content-Type: multipart/form-data" \
      -F "messages=Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5." \
      -F "max_tokens=32" \
      -F "stream=false"
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service faqgen-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with the ChatQnA workload.

## Values

| Key               | Type   | Default                                 | Description                                                                            |
| ----------------- | ------ | --------------------------------------- | -------------------------------------------------------------------------------------- |
| image.repository  | string | `"opea/faqgen"`                         |                                                                                        |
| service.port      | string | `"8888"`                                |                                                                                        |
| tgi.LLM_MODEL_ID  | string | `"meta-llama/Meta-Llama-3-8B-Instruct"` | Models id from https://huggingface.co/, or predownloaded model directory               |
| global.monitoring | bool   | `false`                                 | Enable usage metrics for the service components. See ../monitoring.md before enabling! |
