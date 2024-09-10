# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_library",
)

licenses(["notice"])

package(default_visibility = ["//visibility:public"])

envoy_cc_library(
    name = "config",
    srcs = ["config.cc"],
    hdrs = ["config.h"],
    repository = "@envoy",
    linkopts = [
        "-lopenvino",
        "-lopenvino_c",
        "-lopenvino_ir_frontend",
    ],
    deps = [
        "//source/extensions/common/inference:inference_runtime_interface",
        "@envoy//envoy/thread_local:thread_local_interface",
    ],
)
