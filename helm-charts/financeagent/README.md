# FinanceAgent

Helm chart for deploying FinanceAgent example. It demonstrates how agent works, using prepared data and questions. See [FinanceAgent Overview](https://github.com/opea-project/GenAIExamples/tree/main/FinanceAgent#overview) for the details.

FinanceAgent depends on the following subcharts:

- [agent](../common/agent/README.md)
- [llm-uservice](../common/llm-uservice/README.md)
- [vllm](../common/vllm/README.md)
- [tei](../common/tei/README.md)
- [redis-vector-db](../common/redis-vector-db/README.md)
- [data-prep](../common/data-prep/README.md)

Check the [Chart.yaml](./Chart.yaml) file for how these subcharts are used.

Agent usually requires larger models to perform better, we used `meta-llama/Llama-3.3-70B-Instruct` for test, which requires 4x Gaudi devices for local deployment.

## Deploy

The Deployment includes downloading tools and prompts for the agents, and data ingestion for testing.

### Prerequisites

A volume is required to put tools and prompts used by agent.

We'll use hostPath in this README, which is convenient for single worker node deployment. PVC is recommended in a bigger cluster. If you want to use a PVC, comment out the `toolHostPath` and replace with `toolPVC` in the `values.yaml`.

Create the directory `/mnt/tools/financeagent` in the worker node, which is the default in `values.yaml`. We use the same directory for all 3 agents for easy configuration.

```
sudo mkdir /mnt/tools/financeagent
sudo chmod 777 /mnt/tools/financeagent
```

Download prompts, tools and the configuration to `/mnt/tools/financeagent`

```
# prompts used by 3 agents
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/prompts/finqa_prompt.py -O /mnt/tools/financeagent/finqa_prompt.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/prompts/research_prompt.py -O /mnt/tools/financeagent/research_prompt.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/prompts/supervisor_prompt.py -O /mnt/tools/financeagent/supervisor_prompt.py

# tools and configurations used by the agents
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/finqa_agent_tools.yaml -O /mnt/tools/financeagent/finqa_agent_tools.yaml
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/finqa_tools.py -O /mnt/tools/financeagent/finqa_tools.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/research_agent_tools.yaml -O /mnt/tools/financeagent/research_agent_tools.yaml
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/research_tools.py -O /mnt/tools/financeagent/research_tools.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/supervisor_agent_tools.yaml -O /mnt/tools/financeagent/supervisor_agent_tools.yaml
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/supervisor_tools.py -O /mnt/tools/financeagent/supervisor_tools.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/sum_agent_tools.py -O /mnt/tools/financeagent/sum_agent_tools.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/redis_kv.py -O /mnt/tools/financeagent/redis_kv.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/FinanceAgent/tools/utils.py -O /mnt/tools/financeagent/utils.py
```

### Deploy with Helm chart

Deploy everything on Gaudi enabled Kubernetes cluster:

If you want to try with latest version, use `helm pull oci://ghcr.io/opea-project/charts/financeagent --version 0-latest --untar`

```
export HF_TOKEN="YourOwnToken"
export FINNHUB_API_KEY="YourOwnToken"
export FINANCIAL_DATASETS_API_KEY="YourOwnToken"
helm pull oci://ghcr.io/opea-project/charts/financeagent --untar
helm install financeagent financeagent -f financeagent/gaudi-values.yaml \
--set global.HF_TOKEN=${HF_TOKEN} \
--set research-agent.FINNHUB_API_KEY=${FINNHUB_API_KEY}  \
--set research-agent.FINANCIAL_DATASETS_API_KEY=${FINANCIAL_DATASETS_API_KEY}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

### Ingest data for RAG

Ingest data used by RAG.

```
ip_address=$(kubectl get svc financeagent-data-prep -o jsonpath='{.spec.clusterIP})
curl "http://${ip_address}:6007/v1/dataprep/ingest" -X POST  -H "Content-Type: application/x-www-form-urlencoded" \
  -d 'link_list=%5B%22https%3A%2F%2Fwww.fool.com%2Fearnings%2Fcall-transcripts%2F2025%2F03%2F06%2Fcostco-wholesale-cost-q2-2025-earnings-call-transc%2F%22%2C%22https%3A%2F%2Fwww.fool.com%2Fearnings%2Fcall-transcripts%2F2025%2F03%2F07%2Fgap-gap-q4-2024-earnings-call-transcript%2F%22%5D'
```

### Verify the workload through curl command

Run the command `kubectl port-forward svc/financeagent-supervisor 9090:9090` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9090/v1/chat/completions \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"role": "user", "messages": "Can you summarize Costco 2025 Q2 earnings call?"}'
```
