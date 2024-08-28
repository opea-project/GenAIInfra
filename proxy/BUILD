load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_binary",
)
load("@rules_pkg//:pkg.bzl", "pkg_tar")

exports_files(["LICENSE"])

ISTIO_EXTENSIONS = [
    "@proxy//source/extensions/common/workload_discovery:api_lib",  # Experimental: WIP
    "@proxy//source/extensions/filters/http/alpn:config_lib",
    "@proxy//source/extensions/filters/http/istio_stats",
    "@proxy//source/extensions/filters/http/peer_metadata:filter_lib",
    "@proxy//source/extensions/filters/network/metadata_exchange:config_lib",
]

OPEA_EXTENSIONS = [
    "@envoy//contrib/hyperscan/matching/input_matchers/source:config",
    "@envoy//contrib/hyperscan/regex_engines/source:config",
]

envoy_cc_binary(
    name = "envoy",
    repository = "@envoy",
    deps = ISTIO_EXTENSIONS + OPEA_EXTENSIONS + [
        "@envoy//source/exe:envoy_main_entry_lib",
    ],
)

pkg_tar(
    name = "envoy_tar",
    srcs = [":envoy"],
    extension = "tar.gz",
    mode = "0755",
    package_dir = "/usr/local/bin/",
    tags = ["manual"],
)
