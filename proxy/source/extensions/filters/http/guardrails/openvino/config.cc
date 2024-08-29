#include "source/extensions/filters/http/guardrails/openvino/config.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {
namespace OpenVino {

CompiledModelThreadLocal::CompiledModelThreadLocal(ov::Core core,
                                                   std::shared_ptr<ov::Model> model) {
  TRY_ASSERT_MAIN_THREAD { compiled_model_ = core.compile_model(model, "CPU"); }
  END_TRY CATCH(const ov::Exception& e,
                { throw EnvoyException(fmt::format("Failed to compile model: {}", e.what())); });
}

Session::Session(Runtime* runtime) : runtime_(runtime) {}

bool Session::classify(const std::string& input) {
  ov::InferRequest request = runtime_->compiledModel().create_infer_request();
  std::string& i = const_cast<std::string&>(input);
  ov::Tensor tensor(ov::element::string, ov::Shape{1}, &i);
  request.set_input_tensor(tensor);

  request.start_async();
  request.wait();

  ov::Tensor output = request.get_output_tensor();
  const float* result = output.data<const float>();
  ENVOY_LOG(debug, "Inference output: {}", *result);

  return *result > runtime_->threshold();
}

Runtime::Runtime(const std::string& model_path, double threshold, ThreadLocal::SlotAllocator& tls)
    : threshold_(threshold),
      tls_(ThreadLocal::TypedSlot<CompiledModelThreadLocal>::makeUnique(tls)) {
  // Add OpenVINO Tokenizers extension.
  TRY_ASSERT_MAIN_THREAD { core_.add_extension("/usr/lib/libopenvino_tokenizers.so"); }
  END_TRY CATCH(const ov::Exception& e, {
    throw EnvoyException(fmt::format("Failed to add OpenVINO Tokenizers: {}", e.what()));
  });
  TRY_ASSERT_MAIN_THREAD { model_ = core_.read_model(model_path); }
  END_TRY CATCH(const ov::Exception& e,
                { throw EnvoyException(fmt::format("Failed to read model: {}", e.what())); });
  RELEASE_ASSERT(model_ != nullptr, "Invalid model");

  tls_->set([this](Event::Dispatcher&) {
    return std::make_shared<CompiledModelThreadLocal>(core_, model_);
  });
}

SessionPtr Runtime::createSession() { return std::make_unique<Session>(this); }

} // namespace OpenVino
} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
