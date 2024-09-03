# Copyright (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_test",
)

licenses(["notice"])

package(default_visibility = ["//visibility:public"])

envoy_cc_test(
    name = "filter_test",
    srcs = ["filter_test.cc"],
    data = [
        "//test/extensions/filters/http/guardrails/test_data:model",
    ],
    repository = "@envoy",
    deps = [
        "//source/extensions/filters/http/guardrails:guardrails_lib",
        "@envoy//test/mocks/server:factory_context_mocks",
        "@envoy//test/test_common:utility_lib",
    ],
)
