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
        "@envoy//envoy/thread_local:thread_local_interface",
    ],
)
