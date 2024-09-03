# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_library",
)

licenses(["notice"])

package(default_visibility = ["//visibility:public"])

envoy_cc_library(
    name = "inference_runtime_interface",
    hdrs = ["inference_runtime.h"],
    repository = "@envoy",
    deps = ["@envoy//envoy/common:pure_lib"],
    alwayslink = 1,
)

envoy_cc_library(
    name = "utility_lib",
    srcs = ["utility.cc"],
    hdrs = ["utility.h"],
    repository = "@envoy",
    deps = [
        ":inference_runtime_interface",
        "//source/extensions/inference_runtime/openvino:config",
        "@envoy//envoy/server:factory_context_interface",
    ],
)
