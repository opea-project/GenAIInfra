# Model Files

This directory contains YAML configuration files for various AI models designed to run on Kubernetes clusters. These files define the specifications, arguments, and resource profiles required for deploying and running the models efficiently.

## Benchmarking
The parameters for the models were determined using the KubeAI benchmarking tool benchmark_serving.py The benchmarking script can be found [here](https://github.com/substratusai/kubeai/blob/main/benchmarks/chat-py/benchmark_serving.py).

The following arguments were used during benchmarking:
- `--request-rate=800`
- `--max-concurrency=800`
- `--num-prompts=8000`
- `--max-conversations=800`

These parameters were chosen to optimize the model's performance in terms of throughput.

## Additional Notes
- The `cacheProfile` is set to `default`.
- The `targetRequests` value matches the `max-num-seqs` (batch size).

For more details, refer to the individual YAML files in this directory.
