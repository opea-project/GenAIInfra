# AgentQnA

Helm chart for deploying AgentQnA example. It demonstrates how agent works, using prepared data and questions. See [AgentQnA overview](https://github.com/opea-project/GenAIExamples/tree/main/AgentQnA#overview) for the details.

Using different datasets, models and questions may get different results.

Agent usually requires larger models to perform better, we used Llama-3.3-70B-Instruct for test, which requires 4x Gaudi devices for local deployment.

With helm chart, we also provided option with smaller model (Meta-Llama-3-8B-Instruct) with compromised performance on Xeon CPU only environment for you to try.

## Deploy

The Deployment includes preparing tools and SQL data.

### Prerequisites

A volume is required to put tools configuration used by agent, and the database data used by sqlagent.

We'll use hostPath in this readme, which is convenient for single worker node deployment. PVC is recommended in a bigger cluster. If you want to use a PVC, comment out the `toolHostPath` and replace with `toolPVC` in the `values.yaml`.

Create the directory `/mnt/tools` in the worker node, which is the default in `values.yaml`. We use the same directory for all 3 agents for easy configuration.

```
sudo mkdir /mnt/tools
sudo chmod 777 /mnt/tools
```

Download tools and the configuration to `/mnt/tools`

```
# tools used by supervisor
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/tools/supervisor_agent_tools.yaml -O /mnt/tools/supervisor_agent_tools.yaml
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/tools/tools.py -O /mnt/tools/tools.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/tools/pycragapi.py -O /mnt/tools/pycragapi.py

# tools used by rag agent
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/tools/worker_agent_tools.yaml -O /mnt/tools/worker_agent_tools.yaml
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/tools/worker_agent_tools.py -O /mnt/tools/worker_agent_tools.py
```

Download the `sqlite` database binary file

```
wget https://raw.githubusercontent.com/lerocha/chinook-database/refs/heads/master/ChinookDatabase/DataSources/Chinook_Sqlite.sqlite -O /mnt/tools/Chinook_Sqlite.sqlite
```

### Deploy with Helm chart

Deploy everything on Gaudi enabled Kubernetes cluster:

If you want to try with latest version, use `helm pull oci://ghcr.io/opea-project/charts/agentqna --version 0-latest --untar`

```
export HUGGINGFACEHUB_API_TOKEN="YourOwnToken"
helm pull oci://ghcr.io/opea-project/charts/agentqna --untar
helm install agentqna agentqna -f agentqna/gaudi-values.yaml --set global.HUGGINGFACEHUB_API_TOKEN=${HUGGINGFACEHUB_API_TOKEN}
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all pods are running.

### Ingest data for RAG

Ingest data used by RAG.

```
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/retrieval_tool/index_data.py -O /mnt/tools/index_data.py
wget https://raw.githubusercontent.com/opea-project/GenAIExamples/refs/heads/main/AgentQnA/example_data/test_docs_music.jsonl -O /mnt/tools/test_docs_music.jsonl
host_ip=$(kubectl get svc -o jsonpath="{.items[].spec.clusterIP}" --selector app.kubernetes.io/name=data-prep)
python3 index_data.py --filedir /mnt/tools --filename test_docs_music.jsonl --host_ip $host_ip
```

### Verify the workload through curl command

Run the command `kubectl port-forward svc/agentqna-supervisor 9090:9090` to expose the service for access.

Open another terminal and run the following command to verify the service if working:

```console
curl http://localhost:9090/v1/chat/completions \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"messages": "How many albums does Iron Maiden have?"}'
```
