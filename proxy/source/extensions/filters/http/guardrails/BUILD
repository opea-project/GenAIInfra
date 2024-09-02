load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_library",
)

licenses(["notice"])

package(default_visibility = ["//visibility:public"])

envoy_cc_library(
    name = "guardrails_lib",
    srcs = ["filter.cc"],
    hdrs = ["filter.h"],
    repository = "@envoy",
    deps = [
        "//source/extensions/common/inference:utility_lib",
        "@envoy//envoy/server:filter_config_interface",
        "@envoy//source/common/buffer:buffer_lib",
        "@envoy//source/common/json:json_loader_lib",
    ],
)

envoy_cc_library(
    name = "config",
    srcs = ["config.cc"],
    hdrs = ["config.h"],
    repository = "@envoy",
    deps = [
        ":config_cc_proto",
        ":guardrails_lib",
        "@envoy//envoy/registry",
        "@envoy//source/extensions/filters/http/common:factory_base_lib",
    ],
)

cc_proto_library(
    name = "config_cc_proto",
    deps = ["proto"],
)

proto_library(
    name = "proto",
    srcs = ["config.proto"],
)
