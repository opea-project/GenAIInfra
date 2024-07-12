# web-retriever

Helm chart for deploying Web Retriever microservice.

Web retriever depends on tei, you should set TEI_EMBEDDING_ENDPOINT endpoints before start.

## Installing the Chart

To install the chart, run the following:

```console
export TEI_EMBEDDING_ENDPOINT="http://tei"
export GOOGLE_API_KEY="yourownkey"
export GOOGLE_CSE_ID="yourownid"
helm install web-retriever web-retriever --set TEI_EMBEDDING_ENDPOINT=${TEI_EMBEDDING_ENDPOINT} \
--set GOOGLE_API_KEY=${GOOGLE_API_KEY} \
--set GOOGLE_CSE_ID=${GOOGLE_CSE_ID}
```

## Verify

Use port-forward to access it from localhost.

```console
kubectl port-forward service/web-retriever 1234:7077 &
your_embedding=$(python -c "import random; embedding = [random.uniform(-1, 1) for _ in range(768)]; print(embedding)")
curl http://localhost:1234/v1/web_retrieval \
  -X POST \
  -d "{\"text\":\"What is OPEA?\",\"embedding\":${your_embedding}}" \
  -H 'Content-Type: application/json'
```

## Values

| Key                    | Type   | Default                              | Description |
| ---------------------- | ------ | ------------------------------------ | ----------- |
| image.repository       | string | `"opea/web-retriever-chroma:latest"` |             |
| service.port           | string | `"7077"`                             |             |
| TEI_EMBEDDING_ENDPOINT | string | `""`                                 |             |
| GOOGLE_API_KEY         | string | `""`                                 |             |
| GOOGLE_CSE_ID          | string | `""`                                 |             |
