# txt2img

Helm chart for deploying txt2img service.

## Installing the Chart

To install the chart, run the following:

```console
cd GenAIInfra/helm-charts/
./update_dependency.sh
helm dependency update txt2img
export HFTOKEN="insert-your-huggingface-token-here"
export MODELDIR="/mnt/opea-models"
# To run on Xeon
helm install txt2img txt2img --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR}
# To run on Gaudi
#helm install txt2img txt2img --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set global.modelUseHostPath=${MODELDIR} -f txt2img/gaudi-values.yaml
```

### IMPORTANT NOTE

1. Make sure your `MODELDIR` exists on the node where your workload is schedueled so you can cache the downloaded model for next time use. Otherwise, set `global.modelUseHostPath` to 'null' if you don't want to cache the model.

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Curl command and UI are the two options that can be leveraged to verify the result.

### Verify the workload through curl command

Then run the command `kubectl port-forward svc/txt2img-text2image 9379:9379` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9379/v1/text2image \
    -XPOST \
    -d '{"prompt":"An astronaut riding a green horse", "num_images_per_prompt":1}' \
    -H 'Content-Type: application/json'
```

### Verify the workload through UI

The UI has already been installed via the Helm chart. To access it, use the external IP of one your Kubernetes node along with the NGINX port. You can find the NGINX port using the following command:

```bash
export port=$(kubectl get service txt2img-nginx --output='jsonpath={.spec.ports[0].nodePort}')
echo $port
```

Open a browser to access `http://<k8s-node-ip-address>:${port}` to play with the txt2img workload.

## Values

| Key              | Type   | Default                                         | Description                                                              |
| ---------------- | ------ | ----------------------------------------------- | ------------------------------------------------------------------------ |
| text2image.MODEL | string | `"stable-diffusion-v1-5/stable-diffusion-v1-5"` | Models id from https://huggingface.co/, or predownloaded model directory |
