Let's use this file for a discussion of how to set up the Helm probe and resource in common helm chart, e.g. tgi, tei, etc.

Below is the proposal from @eero-t :

[1] What's IMHO needed to fix things for probe timings:

- Common components' value files include different probe timing sections for CPU and for accelerators

- Their deployment templates select one based on .Values.accelDevice value (empty for CPU)

- All <device>-values.yaml files set appropriate <subchart>.accelDevice value (not just ChatQnA)

- GMC device variant manifests are generated for all relevant components, not just TGI

(I don't think probe timings would need to be fine-tuned based on which model is used.)

[2] What's IMHO needed to fix resource requests:

- Current sub-optimal component arguments are optimized, and resulting resource requirements are profiled, for all relevant models

  - For example on SPR, CPU TGI data type can be set to bfloat16, which halves its memory usage
  - Observability support + suitable Grafana dashboards will help with profiling

- Instead of subcomponent model & corresponding resources being specified in top-level chart, helm install command uses suitable model+resource+args file from given component, like this:
  -f common/tgi/gaudi/neural-chat-7b.yaml
  -f common/teirerank/gaudi/bge-reranker-base.yaml
  -f common/tei/cpu/bge-base.yaml
  -f common/data-prep/gpu/values.yaml
  (These would provide values with subchart prefix/heading so they can be used from top-level charts)

- There could also be a global option for ignoring (CPU side) resource requests that can be used when things need to be re-profiled

- GMC applies resource specs generated from these, when user changes model

If there are combinations of above which are common between different top-level charts, I would suggest Makefile merging relevant ones to common files (e.g. -f common/gaudi-model-defaults.yaml), to avoid duplicating things that may need to be updated whenever args, model, or image versions get updated.
