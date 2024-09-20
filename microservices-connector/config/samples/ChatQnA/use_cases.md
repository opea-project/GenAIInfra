# ChatQnA Use Cases in Kubernetes Cluster via GMC

This document outlines the deployment process for a ChatQnA application utilizing the [GenAIComps](https://github.com/opea-project/GenAIComps.git) microservice pipeline components on Intel Xeon server and Gaudi machines.

The ChatQnA Service leverages a Kubernetes operator called genai-microservices-connector(GMC). GMC supports connecting microservices to create pipelines based on the specification in the pipeline yaml file in addition to allowing the user to dynamically control which model is used in a service such as an LLM or embedder. The underlying pipeline language also supports using external services that may be running in public or private cloud elsewhere.

Install GMC in your Kubernetes cluster, if you have not already done so, by following the steps in Section "Getting Started" at [GMC Install](https://github.com/opea-project/GenAIInfra/tree/main/microservices-connector). Soon as we publish images to Docker Hub, at which point no builds will be required, simplifying install.

The ChatQnA application is defined as a Custom Resource (CR) file that the above GMC operator acts upon. It first checks if the microservices listed in the CR yaml file are running, if not starts them and then proceeds to connect them. When the ChatQnA RAG pipeline is ready, the service endpoint details are returned, letting you use the application. Should you use "kubectl get pods" commands you will see all the component microservices, in particular `embedding`, `retriever`, `rerank`, and `llm`.

## Using prebuilt images

The ChatQnA uses the below prebuilt images if you choose a Xeon deployment

- embedding: opea/embedding-tei:latest
- retriever: opea/retriever-redis:latest
- reranking: opea/reranking-tei:latest
- llm: opea/llm-tgi:latest
- dataprep-redis: opea/dataprep-redis:latest
- tei_xeon_service: ghcr.io/huggingface/text-embeddings-inference:cpu-1.5
- tei_embedding_service: ghcr.io/huggingface/text-embeddings-inference:cpu-1.5
- tgi-service: ghcr.io/huggingface/text-generation-inference:sha-e4201f4-intel-cpu
- redis-vector-db: redis/redis-stack:7.2.0-v9

Should you desire to use the Gaudi accelerator, two alternate images are used for the embedding and llm services.
For Gaudi:

- tei-embedding-service: ghcr.io/huggingface/tei-gaudi:synapse_1.16
- tgi-service: ghcr.io/huggingface/tgi-gaudi:2.0.1

## Deploy ChatQnA pipeline

There are 3 use cases for ChatQnA example:

- General ChatQnA with preset RAG data
- ChatQnA with data preparation which supports that the user can upload RAG data online via dataprep microservice
- ChatQnA supports multiple LLM models which can be switched in runtime

### General ChatQnA with preset RAG data

This involves deploying the ChatQnA custom resource. You can use `chatQnA_xeon.yaml` or if you have a Gaudi cluster, you could use `chatQnA_gaudi.yaml`.

1. Create namespace and deploy application

   ```sh
   kubectl create ns chatqa
   kubectl apply -f $(pwd)/chatQnA_xeon.yaml
   ```

2. GMC will reconcile the ChatQnA custom resource and get all related components/services ready. Check if the service up.

   ```sh
   kubectl get service -n chatqa
   ```

3. Retrieve the application access URL

   ```sh
   kubectl get gmconnectors.gmc.opea.io -n chatqa
   NAME     URL                                                      READY     AGE
   chatqa   http://router-service.chatqa.svc.cluster.local:8080      9/0/9     3m
   ```

4. Deploy a client pod to test the application

   ```sh
   kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity
   ```

5. Access the application using the above URL from the client pod

   ```sh
   export CLIENT_POD=$(kubectl get pod -n chatqa -l app=client-test -o jsonpath={.items..metadata.name})
   export accessUrl=$(kubectl get gmc -n chatqa -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n chatqa -- curl -s --no-buffer $accessUrl  -X POST  -d '{"text":"What is the revenue of Nike in 2023?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json'
   ```

6. Perhaps you want to try another LLM model? Just modify the application custom resource to use another LLM model

   Should you, for instance, want to change the LLM model you are using in the ChatQnA pipeline, just edit the custom resource file.
   For example, to use Llama-2-7b-chat-hf make the following edit:

   ```yaml
   - name: Tgi
     internalService:
       serviceName: tgi-service-m
       config:
         LLM_MODEL_ID: Llama-2-7b-chat-hf
   ```

7. Apply the change

   ```
   kubectl apply -f $(pwd)/chatQnA_xeon.yaml
   ```

8. Check that the tgi-svc-deployment has been changed to use the new LLM Model

   ```sh
   kubectl get deployment tgi-service-m-deployment -n chatqa -o jsonpath="{.spec.template.spec.containers[*].env[?(@.name=='LLM_MODEL_ID')].value}"
   ```

9. Access the updated pipeline using the same URL from above using the client pod

   ```sh
   kubectl exec "$CLIENT_POD" -n chatqa -- curl -s --no-buffer $accessUrl -X POST -d '{"text":"What are the key features of Intel Gaudi?","parameters":{"max_new_tokens":17, "do_sample": true}}' -H 'Content-Type: application/json'
   ```

> [NOTE]

You can remove your ChatQnA pipeline by executing standard Kubernetes kubectl commands to remove a custom resource. Verify it was removed by executing kubectl get pods in the chatqa namespace.

### ChatQnA with data preparation

This involves deploying the ChatQnA custom resource. You can use `chatQnA_dataprep_xeon.yaml` or if you have a Gaudi cluster, you could use `chatQnA_dataprep_gaudi.yaml`.

1. Create namespace and deploy application

   ```sh
   kubectl create ns chatqa
   kubectl apply -f $(pwd)/chatQnA_dataprep_xeon.yaml
   ```

2. GMC will reconcile the ChatQnA custom resource and get all related components/services ready. Check if the service up.

   ```sh
   kubectl get service -n chatqa
   ```

3. Retrieve the application access URL

   ```sh
   kubectl get gmconnectors.gmc.opea.io -n chatqa
   NAME     URL                                                      READY     AGE
   chatqa   http://router-service.chatqa.svc.cluster.local:8080      10/0/10    3m
   ```

   > [NOTE]
   > Comparing with `General ChatQnA with preset RAG data`, there should be `10` microservices, the extra one is the microservice of `dataprep`.

4. Deploy a client pod to test the application

   ```sh
   kubectl create deployment client-test -n chatqa --image=python:3.8.13 -- sleep infinity
   ```

5. Upload the RAG data from internet via microservice `dataprep`

   ```sh
   export CLIENT_POD=$(kubectl get pod -n chatqa -l app=client-test -o jsonpath={.items..metadata.name})
   export accessUrl=$(kubectl get gmc -n chatqa -o jsonpath="{.items[?(@.metadata.name=='chatqa')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n chatqa -- curl -s --no-buffer "$accessUrl/dataprep" -F 'link_list=["https://raw.githubusercontent.com/opea-project/GenAIInfra/main/microservices-connector/test/data/gaudi.txt"]' -H "Content-Type: multipart/form-data"
   ```

6. Access the application using the above URL from the client pod

   ```sh
   kubectl exec "$CLIENT_POD" -n chatqa -- curl -s --no-buffer $accessUrl  -X POST  '{"text":"What are the key features of Intel Gaudi?","parameters":{"max_new_tokens":100, "do_sample": true}}' -H 'Content-Type: application/json'
   ```

   > [NOTE]
   > You can remove your ChatQnA pipeline by executing standard Kubernetes kubectl commands to remove a custom resource. Verify it was removed by executing kubectl get pods in the chatqa namespace.

### ChatQnA supports multiple LLM models

This involves deploying the ChatQnA custom resource. You can use `chatQnA_switch_xeon.yaml` or if you have a Gaudi cluster, you could use `chatQnA_switch_gaudi.yaml`. Moreover, this use case contains 2 LLM models: `Intel/neural-chat-7b-v3-3` and `meta-llama/CodeLlama-7b-hf`.

1. Create namespace and deploy application

   ```sh
   kubectl create ns switch
   kubectl apply -f $(pwd)/chatQnA_switch_xeon.yaml
   ```

2. GMC will reconcile the ChatQnA custom resource and get all related components/services ready. Check if the service up.

   ```sh
   kubectl get service -n switch
   ```

3. Retrieve the application access URL

   ```sh
   kubectl get gmconnectors.gmc.opea.io -n switch
   NAME     URL                                                   READY     AGE
   switch   http://router-service.switch.svc.cluster.local:8080   15/0/15   83s
   ```

   > [NOTE]
   > Comparing with `General ChatQnA with preset RAG data`, there should be `15` microservices, the extra are the microservices for different embedding models and LLM models.

4. Deploy a client pod to test the application

   ```sh
   kubectl create deployment client-test -n switch --image=python:3.8.13 -- sleep infinity
   ```

5. Access the application using the above URL from the client pod by using LLM model `Intel/neural-chat-7b-v3-3`

   ```sh
   export CLIENT_POD=$(kubectl get pod -n switch -l app=client-test -o jsonpath={.items..metadata.name})
   export accessUrl=$(kubectl get gmc -n switch -o jsonpath="{.items[?(@.metadata.name=='switch')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n switch -- curl -s --no-buffer $accessUrl  -X POST  -d '{"text":"What are the key features of Intel Gaudi?", "model-id":"intel", "embedding-model-id":"small", "parameters":{"max_new_tokens":50, "do_sample": true}}' -H 'Content-Type: application/json'
   ```

6. Access the application using the above URL from the client pod by using LLM model `meta-llama/CodeLlama-7b-hf`

   ```sh
   export CLIENT_POD=$(kubectl get pod -n switch -l app=client-test -o jsonpath={.items..metadata.name})
   export accessUrl=$(kubectl get gmc -n switch -o jsonpath="{.items[?(@.metadata.name=='switch')].status.accessUrl}")
   kubectl exec "$CLIENT_POD" -n switch -- curl -s --no-buffer $accessUrl  -X POST  -d '{"text":"What are the key features of Intel Gaudi?", "model-id":"llama", "embedding-model-id":"small", "parameters":{"max_new_tokens":50, "do_sample": true}}' -H 'Content-Type: application/json'
   ```

   > [NOTE]
   > Showing as above, user can switch the LLM models in runtime by changing the request body, such as adding `"model-id":"llama"` in request body to use the Llama model or changing it into `"model-id":"intel"` to use the Intel model.
