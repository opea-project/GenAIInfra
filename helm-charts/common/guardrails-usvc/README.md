# guardrails-usvc

Helm chart for deploying Guardrails microservice.

## Installing the chart

`guardrails-usvc` depends on the following inference backend services:

- TGI: please refer to [tgi](../tgi) chart for more information

### Use Meta Llama Guard models(default):

First, you need to install `tgi` helm chart using the model `meta-llama/Meta-Llama-Guard-2-8B`.

After you've deployed the dependent chart successfully, please run `kubectl get svc` to get the backend inference service endpoint, e.g. `http://tgi`.

To install the `guardrails-usvc` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/guardrails-usvc
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"
export SAFETY_GUARD_ENDPOINT="http://tgi"
export SAFETY_GUARD_MODEL_ID="meta-llama/Meta-Llama-Guard-2-8B"
export GUARDRAILS_BACKEND="LLAMA"
helm install guardrails-usvc . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set SAFETY_GUARD_ENDPOINT=${SAFETY_GUARD_ENDPOINT} --set SAFETY_GUARD_MODEL_ID=${SAFETY_GUARD_MODEL_ID} --set GUARDRAILS_BACKEND=${GUARDRAILS_BACKEND} --wait
```

### Use Allen Institute AI's WildGuard models:

First, you need to install `tgi` helm chart using the model `allenai/wildguard`.

After you've deployed the dependent chart successfully, please run `kubectl get svc` to get the backend inference service endpoint, e.g. `http://tgi`.

To install the `guardrails-usvc` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/guardrails-usvc
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"
export SAFETY_GUARD_ENDPOINT="http://tgi"
export SAFETY_GUARD_MODEL_ID="allenai/wildguard"
export GUARDRAILS_BACKEND="WILD"
helm install guardrails-usvc . --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --set SAFETY_GUARD_ENDPOINT=${SAFETY_GUARD_ENDPOINT} --set SAFETY_GUARD_MODEL_ID=${SAFETY_GUARD_MODEL_ID} --set GUARDRAILS_BACKEND=${GUARDRAILS_BACKEND} --wait
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/guardrails-usvc 9090:9090` to expose the guardrails-usvc service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9090/v1/guardrails \
    -X POST \
    -d '{"text":"How do you buy a tiger in the US?","parameters":{"max_new_tokens":32}}' \
    -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                              | Description                                                     |
| ------------------------------- | ------ | ------------------------------------ | --------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                                 | Your own Hugging Face API token                                 |
| image.repository                | string | `"opea/guardrails-usvc"`             |                                                                 |
| service.port                    | string | `"9090"`                             |                                                                 |
| SAFETY_GUARD_ENDPOINT           | string | `""`                                 | LLM endpoint                                                    |
| SAFETY_GUARD_MODEL_ID           | string | `"meta-llama/Meta-Llama-Guard-2-8B"` | Model ID for the underlying LLM service is using                |
| GUARDRAIL_BACKEND               | string | `"LLAMA"`                            | different gaurdrail model family to use, one of `LLAMA`, `WILD` |
