# llm-uservice

Helm chart for deploying OPEA LLM microservices.

## Installing the chart

`llm-uservice` depends on one of the following inference backend services:

- TGI: please refer to [tgi](../tgi) chart for more information

- vLLM: please refer to [vllm](../vllm) chart for more information

First, you need to install one of the dependent chart, i.e. `tgi` or `vllm` helm chart.

After you've deployed the dependent chart successfully, please run `kubectl get svc` to get the backend inference service endpoint, e.g. `http://tgi`, `http://vllm`.

To install the `llm-uservice` chart, run the following:

```console
cd GenAIInfra/helm-charts/common/llm-uservice
helm dependency update
export HFTOKEN="insert-your-huggingface-token-here"
# set backend inferene service endpoint URL
# for tgi
export LLM_ENDPOINT="http://tgi"
# for vllm
# export LLM_ENDPOINT="http://vllm"

# set the same model used by the backend inference service
export LLM_MODEL_ID="Intel/neural-chat-7b-v3-3"

# install llm-textgen with TGI backend
helm install llm-uservice . --set TEXTGEN_BACKEND="TGI" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait

# install llm-textgen with vLLM backend
# helm install llm-uservice . --set TEXTGEN_BACKEND="vLLM" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait

# install llm-docsum with TGI backend
# helm install llm-uservice . --set image.repository="opea/llm-docsum" --set DOCSUM_BACKEND="TGI" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set MAX_INPUT_TOKENS=2048 --set MAX_TOTAL_TOKENS=4096 --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait

# install llm-docsum with vLLM backend
# helm install llm-uservice . --set image.repository="opea/llm-docsum" --set DOCSUM_BACKEND="vLLM" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set MAX_INPUT_TOKENS=2048 --set MAX_TOTAL_TOKENS=4096 --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait

# install llm-faqgen with TGI backend
# helm install llm-uservice . --set image.repository="opea/llm-faqgen" --set FAQGEN_BACKEND="TGI" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait

# install llm-faqgen with vLLM backend
# helm install llm-uservice . --set image.repository="opea/llm-faqgen" --set FAQGEN_BACKEND="vLLM" --set LLM_ENDPOINT=${LLM_ENDPOINT} --set LLM_MODEL_ID=${LLM_MODEL_ID} --set global.HUGGINGFACEHUB_API_TOKEN=${HFTOKEN} --wait
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

Then run the command `kubectl port-forward svc/llm-uservice 9000:9000` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
# for llm-textgen service
curl http://localhost:9000/v1/chat/completions \
  -X POST \
  -d d '{"model": "${LLM_MODEL_ID}", "messages": "What is Deep Learning?", "max_tokens":17}' \
  -H 'Content-Type: application/json'

# for llm-docsum service
curl http://localhost:9000/v1/docsum \
  -X POST \
  -d '{"query":"Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5.", "max_tokens":32, "language":"en"}' \
  -H 'Content-Type: application/json'

# for llm-faqgen service
curl http://localhost:9000/v1/faqgen \
  -X POST \
  -d '{"query":"Text Embeddings Inference (TEI) is a toolkit for deploying and serving open source text embeddings and sequence classification models. TEI enables high-performance extraction for the most popular models, including FlagEmbedding, Ember, GTE and E5.","max_tokens": 128}' \
  -H 'Content-Type: application/json'
```

## Values

| Key                             | Type   | Default                       | Description                                                                      |
| ------------------------------- | ------ | ----------------------------- | -------------------------------------------------------------------------------- |
| global.HUGGINGFACEHUB_API_TOKEN | string | `""`                          | Your own Hugging Face API token                                                  |
| image.repository                | string | `"opea/llm-textgen"`          | one of "opea/llm-textgen", "opea/llm-docsum", "opea/llm-faqgen"                  |
| LLM_ENDPOINT                    | string | `""`                          | backend inference service endpoint                                               |
| LLM_MODEL_ID                    | string | `"Intel/neural-chat-7b-v3-3"` | model used by the inference backend                                              |
| TEXTGEN_BACKEND                 | string | `"tgi"`                       | backend inference engine, only valid for llm-textgen image, one of "TGI", "vLLM" |
| DOCSUM_BACKEND                  | string | `"tgi"`                       | backend inference engine, only valid for llm-docsum image, one of "TGI", "vLLM"  |
| FAQGEN_BACKEND                  | string | `"tgi"`                       | backend inference engine, only valid for llm-faqgen image, one of "TGi", "vLLM"  |
| global.monitoring               | bool   | `false`                       | Service usage metrics                                                            |
