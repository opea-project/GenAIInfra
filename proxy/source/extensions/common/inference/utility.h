#pragma once

#include "envoy/server/factory_context.h"

#include "source/extensions/common/inference/inference_runtime.h"

namespace Envoy {
namespace Extensions {
namespace Common {
namespace Inference {

InferenceRuntimePtr createInferenceRuntime(const std::string& model_path, double threshold,
                                           Server::Configuration::FactoryContext& context);

} // namespace Inference
} // namespace Common
} // namespace Extensions
} // namespace Envoy
