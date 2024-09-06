/*
 * Copyright (C) 2024 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <memory>

#include "envoy/common/pure.h"

namespace Envoy {
namespace Extensions {
namespace Common {
namespace Inference {

class InferenceSession {
public:
  virtual ~InferenceSession() = default;

  virtual bool classify(const std::string& input) PURE;
};

using InferenceSessionPtr = std::unique_ptr<InferenceSession>;

class InferenceRuntime {
public:
  virtual ~InferenceRuntime() = default;

  virtual InferenceSessionPtr createInferenceSession() PURE;
};

using InferenceRuntimePtr = std::unique_ptr<InferenceRuntime>;

} // namespace Inference
} // namespace Common
} // namespace Extensions
} // namespace Envoy
