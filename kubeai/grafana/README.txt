# Files

## Dashboards

- `vllm-scaling.*`: Cluster overview of how much and well vLLM is scaling
- `vllm-details.*`: More detailed per-model and/or per-instance vLLM metrics

## File types

- `*.yaml`: Grafana dashboard configMaps that Grafana will load automatically
- `*.json`: Grafana Dashboard specs from which configMaps are generated
- `*.png`: Screenshots of those dashboards (updated manually)

## Other files

- `convert-dashboard.sh`: convert dashboard `*.json` file to configMap `*.yaml` file
- `README.md`: this file

## Dashboard formats

Dashboard JSON files need to be loaded from the Grafana GUI manually,
their changes can be saved and Grafana maintains update history for
them, but those are lost if Grafana is uninstalled.

Whereas Grafana will automatically load suitably labeled dashboard
configMaps, they persist even if Grafana gets re-installed, and one
can save their updates as JSON files.

Update process for the dashboards is following:

- Apply configMap to K8s so it's visible in Grafana dashboards list
- Update dashboard in Grafana
- Save it as JSON
- Convert JSON file to configMap YAML: `./convert-dashboard.sh *.json`
