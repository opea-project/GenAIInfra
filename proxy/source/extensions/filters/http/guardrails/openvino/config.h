#pragma once

#include "envoy/thread_local/thread_local.h"

#include "openvino/openvino.hpp"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {
namespace OpenVino {

struct CompiledModelThreadLocal : public ThreadLocal::ThreadLocalObject {
  CompiledModelThreadLocal(ov::Core core, std::shared_ptr<ov::Model> model);

  ov::CompiledModel compiled_model_;
};

using CompiledModelThreadLocalPtr = std::unique_ptr<CompiledModelThreadLocal>;

class Runtime;

class Session : Logger::Loggable<Logger::Id::http> {
public:
  Session(Runtime* runtime);

  bool classify(const std::string& input);

protected:
  Runtime* runtime_;
};

using SessionPtr = std::unique_ptr<Session>;

class Runtime {
public:
  Runtime(const std::string& model_path, double threshold, ThreadLocal::SlotAllocator& tls);

  SessionPtr createSession();
  double threshold() const { return threshold_; }
  ov::CompiledModel& compiledModel() { return tls_->get()->compiled_model_; }

protected:
  ov::Core core_;
  std::shared_ptr<ov::Model> model_;
  double threshold_;
  ThreadLocal::TypedSlotPtr<CompiledModelThreadLocal> tls_;
};

using RuntimePtr = std::unique_ptr<Runtime>;

} // namespace OpenVino
} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
