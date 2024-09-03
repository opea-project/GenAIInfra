/*
 * Copyright (C) 2024 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include "envoy/thread_local/thread_local.h"

#include "source/extensions/common/inference/inference_runtime.h"

#include "openvino/openvino.hpp"

namespace Envoy {
namespace Extensions {
namespace InferenceRuntime {
namespace OpenVino {

struct CompiledModelThreadLocal : public ThreadLocal::ThreadLocalObject {
  CompiledModelThreadLocal(ov::Core core, std::shared_ptr<ov::Model> model);

  ov::CompiledModel compiled_model_;
};

using CompiledModelThreadLocalPtr = std::unique_ptr<CompiledModelThreadLocal>;

class InferenceRuntime : public Common::Inference::InferenceRuntime {
public:
  InferenceRuntime(const std::string& model_path, double threshold,
                   ThreadLocal::SlotAllocator& tls);

  Common::Inference::InferenceSessionPtr createInferenceSession() override;
  double threshold() const { return threshold_; }
  ov::CompiledModel& compiledModel() { return tls_->get()->compiled_model_; }

protected:
  ov::Core core_;
  std::shared_ptr<ov::Model> model_;
  double threshold_;
  ThreadLocal::TypedSlotPtr<CompiledModelThreadLocal> tls_;
};

class InferenceSession : public Common::Inference::InferenceSession,
                         Logger::Loggable<Logger::Id::http> {
public:
  InferenceSession(InferenceRuntime* runtime);

  bool classify(const std::string& input) override;

protected:
  InferenceRuntime* runtime_;
};

} // namespace OpenVino
} // namespace InferenceRuntime
} // namespace Extensions
} // namespace Envoy
