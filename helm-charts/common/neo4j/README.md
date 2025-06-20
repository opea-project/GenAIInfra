# neo4j-chart

Helm chart for deploying Neo4j with APOC plugin.

## Install the Chart

To install the chart, run the following:

```console
cd ${YourRepo}/helm-charts/neo4j-chart
helm install neo4j-chart neo4j-chart
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure all the Neo4j pods are running.

Then run the command `kubectl port-forward svc/neo4j 7474:7474` to expose the Neo4j HTTP service for access.

Open another terminal and run the command `curl -I http://127.0.0.1:7474` to access the Neo4j service. The `curl` command should return a response indicating the service is accessible.

## Values

| 🔑 Key                   | 🧩 Type | 🛠️ Default        | 📄 Description                     |
| ------------------------ | ------- | ----------------- | ---------------------------------- |
| `image.repository`       | string  | `"neo4j"`         | The Neo4j image repository         |
| `image.tag`              | string  | `"latest"`        | The Neo4j image tag                |
| `neo4j.username`         | string  | `"neo4j"`         | Default username for Neo4j         |
| `neo4j.password`         | string  | `"password"`      | Default password for Neo4j         |
| `neo4j.ports.http`       | int     | `7474`            | The HTTP port for Neo4j            |
| `neo4j.ports.bolt`       | int     | `7687`            | The Bolt protocol port for Neo4j   |
| `persistence.enabled`    | bool    | `true`            | Enable persistent storage          |
| `persistence.accessMode` | string  | `"ReadWriteOnce"` | Access mode for persistent storage |
| `persistence.size`       | string  | `"10Gi"`          | Size of the persistent storage     |

This `README` provides a concise overview of the Neo4j Helm chart, including installation instructions, verification steps, and a table of configurable values. Adjust any placeholders or values as needed to match your specific deployment environment.
