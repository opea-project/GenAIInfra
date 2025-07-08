# Generating Grafana dashboard Helm chart configMap templates

User would need to manually load Grafana dashboard JSON file, but
when Helm Dashboard chart installs it to Grafana namespace inside a
suitably labeled configMap, Grafana will load it automatically.

Here are the dashboard JSON spec files used as sources for those
configMaps. After dashboard is updated in Grafana, it can be saved
again to a JSON file here, and the corresponding configMap updated
with the provided conversion script.

Usage:

```
cd ../templates/
../json/convert-dashboard.sh ../json/*.json
```
