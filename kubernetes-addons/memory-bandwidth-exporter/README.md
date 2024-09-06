# memory bandwidth exporter

Pod/container grained memory bandwidth exporter provides users memory bandwidth metrics of their running containers. The metrics include llc_occupancy, mbm_local_bytes, mbm_total_bytes, cpu utilization and memory usage, and the metrics have been processed. In addition to container-level metrics, it also provides class-level and socket-level metrics. Users can configure the list of metrics to be collected. It serves as an exporter which can be connected to Promethus-like obserbility tools. And it also can be used as a telementry provider.

Memory bandwidth exporter makes use of state-of-the-art technologies like NRI to build a resource-efficient and well-maintained solution. This solution provides observability to memory bandwidth to OPEA micro-services. It lays the groundwork of better scaling and auto scaling of OPEA. It can also be deployed separately on end user environments, supporting any cases that memory bandwidth metrics are required.

## Setup

### Enable NRI in Containerd

```sh
# download containerd binary, containerd version v1.7.0 or higher is required
$ wget https://github.com/containerd/containerd/releases/download/v1.7.0/containerd-1.7.0-linux-amd64.tar.gz

# stop running containerd
$ sudo systemctl stop containerd

# replace old containerd
$ sudo tar Cxzvf /usr/local containerd-1.7.0-linux-amd64.tar.gz

# enable NRI in containerd
# add an item in /etc/containerd/config.toml
[plugins."io.containerd.nri.v1.nri"]
    disable = false
    disable_connections = false
    plugin_config_path = "/etc/containerd/certs.d"
    plugin_path = "/opt/nri/plugins"
    socket_path = "/var/run/nri/nri.sock"
    config_file = "/etc/nri/nri.conf"

# restart containerd
$ sudo systemctl start containerd
$ sudo systemctl status containerd

# test nri
$ git clone https://github.com/containerd/nri
$ cd nri
$ make
$ ./build/bin/logger -idx 00
```

### Enable RDT

Mount resctrl to the directory `/sys/fs/resctrl`:

```sh
$ sudo mount -t resctrl resctrl /sys/fs/resctrl
```

### Setup memory bandwidth exporter

Before setup, you need to configure the runc hook:

```sh
$ ./config/config.sh
```

#### How to build the binary and setup?

```sh
$ make build
$ sudo ./bin/memory-bandwidth-exporter
# e.g., sudo ./bin/memory-bandwidth-exporter --collector.node.name=<node_name> --collector.container.namespaceWhiteList="calico-apiserver,calico-system,kube-system,tigera-operator"

# get memory bandwidth metrics
$ curl http://localhost:9100/metrics
```

#### How to build the docker image and setup?

```sh
$ make docker.build
$ sudo docker run \
  -e NODE_NAME=<node_name> \
  -e NAMESPACE_WHITELIST="calico-apiserver,calico-system,kube-system,tigera-operator" \
  --mount type=bind,source=/etc/containers/oci/hooks.d/,target=/etc/containers/oci/hooks.d/ \
  --privileged \
  --cgroupns=host \
  --pid=host \
  --mount type=bind,source=/usr/,target=/usr/ \
  --mount type=bind,source=/sys/fs/resctrl/,target=/sys/fs/resctrl/ \
  --mount type=bind,source=/var/run/nri/,target=/var/run/nri/ \
  -d -p 9100:9100 \
  --name=memory-bandwidth-exporter \
  opea/memory-bandwidth-exporter:latest

# get memory bandwidth metrics
$ curl http://localhost:9100/metrics
```

#### How to deploy on the K8s cluster?

Build and push your image to the location specified by `MBE_IMG`, and apply manifest:

```sh
$ make docker.build docker.push MBE_IMG=<some-registry>/opea/memory-bandwidth-exporter:<tag>
$ make change_img MBE_IMG=<some-registry>/opea/memory-bandwidth-exporter:<tag>
# If namespace system does not exist, create it.
$ kubectl create ns system
$ kubectl apply -f config/manifests/memory-bandwidth-exporter.yaml
```

Check the installation result:

```sh
kubectl get pods -n system
NAME                              READY   STATUS    RESTARTS   AGE
memory-bandwidth-exporter-zxhdl   1/1     Running   0          3m
```

get memory bandwidth metrics

```sh
$ curl http://<memory_bandwidth_exporter_container_ip>:9100/metrics
```

#### How to delete binary?

```sh
$ make clean
```

## More flags about memory bandwidth exporter

There are some flags to help users better use memory bandwidth exporter:

```sh
-h, --[no-]help                               Show context-sensitive help (also try --help-long and --help-man).
--collector.node.name=""                      Give node name.
--collector.container.namespaceWhiteList=""   Filter out containers whose namespaces belong to the namespace whitelist, namespaces separated by commas, like "xx,xx,xx".
--collector.container.monTimes=10             Scan the pids of containers created before the exporter starts to prevent the loss of pids.
--collector.container.metrics="all"           Enable container collector metrics.
--collector.class.metrics="none"              Enable class collector metrics.
--collector.node.metrics="none"               Enable node collector metrics.
--web.telemetry-path="/metrics"               Path under which to expose metrics.
--[no-]web.disable-exporter-metrics           Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).
--web.max-requests=40                         Maximum number of parallel scrape requests. Use 0 to disable.
--runtime.gomaxprocs=1                        The target number of CPUs Go will run on (GOMAXPROCS) ($GOMAXPROCS)
--[no-]web.systemd-socket                     Use systemd socket activation listeners instead of port listeners (Linux only).
--web.listen-address=:9100 ...                Addresses on which to expose metrics and web interface. Repeatable for multiple addresses.
--web.config.file=""                          Path to configuration file that can enable TLS or authentication. See: https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
--collector.interval=3s                       memory bandwidth exporter collect metrics interval
--NRIplugin.name="mb-nri-plugin"              Plugin name to register to NRI
--NRIplugin.idx="11"                          Plugin index to register to NRI
--[no-]disableWatch                           Disable watching hook directories for new hooks
--log.level=info                              Only log messages with the given severity or above. One of: [debug, info, warn, error]
--log.format=logfmt                           Output format of log messages. One of: [logfmt, json]
--[no-]version                                Show application version.
```
