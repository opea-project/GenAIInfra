# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_binary",
)

exports_files(["LICENSE"])

ISTIO_EXTENSIONS = [
    "@istio//source/extensions/common/workload_discovery:api_lib",  # Experimental: WIP
    "@istio//source/extensions/filters/http/alpn:config_lib",
    "@istio//source/extensions/filters/http/istio_stats",
    "@istio//source/extensions/filters/http/peer_metadata:filter_lib",
    "@istio//source/extensions/filters/network/metadata_exchange:config_lib",
]

OPEA_EXTENSIONS = [
    "//source/extensions/filters/http/guardrails:config",
]

envoy_cc_binary(
    name = "envoy",
    repository = "@envoy",
    deps = ISTIO_EXTENSIONS + OPEA_EXTENSIONS + [
        "@envoy//source/exe:envoy_main_entry_lib",
    ],
)
