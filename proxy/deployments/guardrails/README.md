# Guardrails

Guardrails is a security feature designed specifically for OPEA RAG pipelines. Guardrails helps eliminate hallucination and toxicity, and build trust from end users.

Guardrails in OPEA Pipeline Proxy supports both rule-based detection and AI-powered content safety. You can leverage rule-based detection to design keyword filter to avoid the leakage of PII (personal identifiable information) or choose a SPAM model to filter SPAMs out.

## Architecture

WIP

## Deployment

The tutorial introduces how to leverage Guardrails and filter out SPAMs in requests on the ingress gateway for a ChatQnA pipeline.

### Configuring the ingress gateway

Before deploying Guardrails, Istio ingress gateway should be setup first. Use the following YAMLs to configure a gateway on port 80 and accept HTTP requests on `/` to the pipeline router.

```yaml
apiVersion: networking.istio.io/v1
kind: Gateway
metadata:
  name: chatqna-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - chatqna-service.com
      port:
        name: http
        number: 80
        protocol: HTTP
```

```yaml
apiVersion: networking.istio.io/v1
kind: VirtualService
metadata:
  name: chatqna-virtual-service
  namespace: istio-system
spec:
  gateways:
    - istio-system/chatqna-gateway
  hosts:
    - chatqna-service.com
  http:
    - match:
        - uri:
            prefix: /
      route:
        - destination:
            host: router-service.chatqa.svc.cluster.local
            port:
              number: 8080
```

### Preparing the model used in Guardrails

OPEA Pipeline Proxy leverages models in OpenVINO IR format. You can convert different models into OpenVINO IR format following [these instructions](https://docs.openvino.ai/2024/openvino-workflow/model-preparation/convert-model-to-ir.html).

A Transformer model is usually consists of a tokenizer and some Transformer blocks. In OpenVINO IR format, the tokenizer and blocks can be merged into a file. You can merge the tokenizer and the model following [these instructions](https://docs.openvino.ai/2024/notebooks/openvino-tokenizers-with-output.html#merge-tokenizer-into-a-model).

In the tutorial, we use [Titeiiko/OTIS-Official-Spam-Model](https://huggingface.co/Titeiiko/OTIS-Official-Spam-Model) for SPAM filtering.

```sh
git clone https://huggingface.co/Titeiiko/OTIS-Official-Spam-Model
ovc OTIS-Official-Spam-Model/model.onnx --output_model OTIS-Official-Spam-Model/model.xml
convert_tokenizer ./OTIS-Official-Spam-Model -o ./OTIS-Official-Spam-Model
```

```python
from openvino import Core, save_model
from openvino_tokenizers import connect_models

core = Core()
tokenizer = core.read_model("OTIS-Official-Spam-Model/openvino_tokenizer.xml")
model = core.read_model("OTIS-Official-Spam-Model/model.xml")
combined_model = connect_models(tokenizer, model)
save_model(combined_model, "OTIS-Official-Spam-Model/combined.xml")
```

### Integrating Guardrails in the ingress gateway

To leverage models, you have to mount them to the OPEA Pipeline Proxy container. Please follow the field [here](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/#KubernetesResourcesSpec) to mount the model into the ingress gateway pod.

You can also use the annotation [here](https://istio.io/latest/docs/reference/config/annotations/#SidecarUserVolume) and [here](https://istio.io/latest/docs/reference/config/annotations/#SidecarUserVolumeMount) to mount volumes in sidecars.

In the tutorial, we mount the SPAM model from the config map to `/model` and insert the Guardrails filter into the ingress gateway.

```sh
kubectl -n istio-system create configmap model --from-file=combined.bin=model.bin --from-file=combined.xml=model.xml
```

```yaml
ingressGateways:
  - name: istio-ingressgateway
    enabled: true
    k8s:
      volumes:
        - name: model
          volumeSource:
            configMap:
              name: model
      volumeMounts:
        - name: model
          mountPath: /model
          readOnly: true
```

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: guardrails
  namespace: istio-system
spec:
  workloadSelector:
    labels:
      istio: ingressgateway
  configPatches:
    - applyTo: HTTP_FILTER
      match:
        context: GATEWAY
        listener:
          portNumber: 8080
          filterChain:
            filter:
              name: envoy.filters.network.http_connection_manager
              subFilter:
                name: envoy.filters.http.router
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.guardrails
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.guardrails.v3.Guardrails
            model_path: /model/combined.xml
            source: REQUEST
            action: ALLOW
```

### Validating the result

Before validation, we have to determine the ingress IP and port. If you are evaluating OPEA Pipeline Proxy in a local environment without any load balancer, you can change the service type from `LoadBalancer` to `NodePort`.

```sh
export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
```

We can then validating the Guardrails with the following `curl` command.

```sh
curl -H "Host:chatqna-service.com" -d "Free money" "http://$INGRESS_HOST:$INGRESS_PORT/"
# Access denied
```
