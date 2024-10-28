# vdms-vector-db

Helm chart for deploying Intel VDMS Vector DB service.

## Install the Chart

To install the chart, run the following:

```bash
cd GenAIInfra/helm-charts/common
helm install vdms-vector-db vdms-vector-db
```

## Verify

To verify the installation, run the command `kubectl get pod` to make sure that vdms-vector-db pod is running.

Then run the command `kubectl port-forward svc/vdms-vector-db 8001:8001` to expose the vdms-vector-db service for access on current host.

Next, verify whether you can reach the `vdms-vector-db` pod. As there are no http services running in `vdms-vector-db` pod, we will verify sanity by making a tcp connection request to the pod and check whether `8001` port is open.

We will use `netcat` command utility for this.

Try running `nc -zv vdms-vector-db 8001`. The command will succeed if vdms-vector-db can accept TCP request and the required port (8001) is open.

## Values

| Key              | Type   | Default            | Description                     |
| ---------------- | ------ | ------------------ | ------------------------------- |
| image.repository | string | `"intellabs/vdms"` |                                 |
| image.tag        | string | `"v2.8.0"`         |                                 |
| service.port     | string | `"8001"`           | The vdms-vector-db service port |
