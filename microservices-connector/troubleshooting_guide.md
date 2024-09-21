# Troubleshooting GMC Custom Resource（CR）

This doc is about identifying issues in GMC CR; A validating webhook has been configured to validate CR fields and it will report all detected errors.
After the CR for GMC pipeline has been deployed, correct the CR if you encounter the following errors:

1. Root node existence

   ```
   The GMCConnector "chatqa" is invalid: spec.nodes: Invalid value: map[string]v1alpha3.Router{"node1":...}: a root node is required
   ```

   In the `spec.nodes` section of the CR, a node with the name ‘root’ is required.

2. StepName validation

   ```
   The GMCConnector "chatqa" is invalid: spec.nodes.root.steps[0].name: Invalid value: v1alpha3.Step{StepName:"Embedding123", Executor:v1alpha3.Executor{NodeName:"", InternalService:v1alpha3.GMCTarget{ServiceName:"embedding-svc", NameSpace:"", Config:map[string]string{"TEI_EMBEDDING_ENDPOINT":"tei-embedding-svc", "endpoint":"/v1/embeddings"}, IsDownstreamService:false}, ExternalService:""}, Data:"", Condition:"", Dependency:"", ServiceURL:""}: invalid step name: Embedding123 for node root
   ```

   In the CR, the value of StepName in the `spec.nodes.<nodeName>.steps[].name` field should be included in the predefined [list](https://github.com/opea-project/GenAIInfra/blob/main/microservices-connector/api/v1alpha3/validating_webhook.go).

3. nodeName existence

   ```
   The GMCConnector "switch" is invalid: spec.nodes.root.steps[0].nodeName: Invalid value: v1alpha3.Step{StepName:"Embedding", Executor:v1alpha3.Executor{NodeName:"node123", InternalService:v1alpha3.GMCTarget{ServiceName:"", NameSpace:"", Config:map[string]string(nil), IsDownstreamService:false}, ExternalService:""}, Data:"", Condition:"", Dependency:"", ServiceURL:""}: node name: node123 in step Embedding does not exist
   ```

   The nodename that is referenced within the `spec.nodes.<nodeName>.steps[].nodeName` field must already be defined in the `spec.nodes` section.

4. serviceName uniqueness

   ```
   The GMCConnector "chatqa" is invalid: spec.nodes.root.steps[1].internalService.serviceName: Invalid value: v1alpha3.Step{StepName:"TeiEmbedding", Executor:v1alpha3.Executor{NodeName:"", InternalService:v1alpha3.GMCTarget{ServiceName:"tei-embedding-svc", NameSpace:"", Config:map[string]string(nil), IsDownstreamService:true}, ExternalService:""}, Data:"", Condition:"", Dependency:"", ServiceURL:""}: service name: tei-embedding-svc in node root already exists
   ```

   The serviceName specified in the `spec.nodes.<nodeName>.steps[].internalService.serviceName` field must be unique and not duplicated with service names in other steps.

5. change log levels
   GMConnectoer's controller support log level management to filter the logs for different purpose.
   To get the current log level, you can send a http GET to GMController like below:
   ```
   curl -X GET "http://localhost:8008/loglevel"
   current log level: info
   ```
   if you want to change the log level for debugging, you can sent a http PUT like below:
   ```
   curl -X PUT "http://localhost:8008/loglevel" -d '{"log_level":"debug"}' -H "Content-Type: application/json"
   log level set to debug
   ```
   the log levels supported by the log system are `debug|info|warn|error|panic|dpanic|panic|tatal`, but current GMC only has the `debug|info|error` logs
