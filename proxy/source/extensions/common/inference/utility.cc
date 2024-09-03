/*
 * Copyright (C) 2024 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

#include "source/extensions/common/inference/utility.h"

#include "source/extensions/inference_runtime/openvino/config.h"

namespace Envoy {
namespace Extensions {
namespace Common {
namespace Inference {

InferenceRuntimePtr createInferenceRuntime(const std::string& model_path, double threshold,
                                           Server::Configuration::FactoryContext& context) {
  InferenceRuntimePtr runtime =
      std::make_unique<Extensions::InferenceRuntime::OpenVino::InferenceRuntime>(
          model_path, threshold, context.serverFactoryContext().threadLocal());
  return runtime;
}

} // namespace Inference
} // namespace Common
} // namespace Extensions
} // namespace Envoy
