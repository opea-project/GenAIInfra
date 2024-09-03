#include "source/extensions/common/inference/utility.h"
#include "source/extensions/filters/http/guardrails/filter.h"

#include "test/mocks/server/factory_context.h"
#include "test/test_common/environment.h"
#include "test/test_common/utility.h"

using testing::_;
using testing::Return;
using testing::ReturnRef;

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

static constexpr absl::string_view ModelPath =
    "/test/extensions/filters/http/guardrails/test_data/model.xml";

std::string GetModelPath() {
  return TestEnvironment::runfilesDirectory("dev_opea_proxy") + std::string(ModelPath);
}

class FilterTest : public testing::Test {
public:
  void setup(std::string& model_path, double threshold, Source source, Action action) {
    EXPECT_CALL(server_factory_context_, threadLocal()).WillOnce(ReturnRef(tls_));
    EXPECT_CALL(context_, serverFactoryContext()).WillOnce(ReturnRef(server_factory_context_));
    Common::Inference::InferenceRuntimePtr runtime =
        Common::Inference::createInferenceRuntime(model_path, threshold, context_);
    config_ = std::make_shared<FilterConfig>(std::move(runtime), source, action);
    filter_ = std::make_shared<Filter>(config_);

    filter_->setDecoderFilterCallbacks(decoded_callbacks_);
    filter_->setEncoderFilterCallbacks(encoded_callbacks_);
  }

protected:
  NiceMock<ThreadLocal::MockInstance> tls_;
  Server::Configuration::MockServerFactoryContext server_factory_context_;
  Server::Configuration::MockFactoryContext context_;
  FilterConfigSharedPtr config_;
  std::shared_ptr<Filter> filter_;
  NiceMock<Http::MockStreamDecoderFilterCallbacks> decoded_callbacks_;
  NiceMock<Http::MockStreamEncoderFilterCallbacks> encoded_callbacks_;
};

TEST_F(FilterTest, RequestAllow) {
  std::string path = GetModelPath();
  setup(path, 0, Source::Request, Action::Allow);

  Buffer::InstancePtr buf(new Buffer::OwnedImpl());
  EXPECT_CALL(decoded_callbacks_, addDecodedData(_, _))
      .WillRepeatedly(Invoke([&](Buffer::Instance& data, bool) { buf->add(data); }));
  EXPECT_CALL(decoded_callbacks_, decodingBuffer()).WillRepeatedly(Return(buf.get()));

  Buffer::OwnedImpl data("{\"text\": \"What a beautiful world!\"}");
  EXPECT_EQ(Http::FilterDataStatus::Continue, filter_->decodeData(data, true));
  buf->drain(buf->length());
  Buffer::OwnedImpl data2("{\"text\": \"Free money!\"}");
  EXPECT_EQ(Http::FilterDataStatus::StopIterationNoBuffer, filter_->decodeData(data2, true));
}

TEST_F(FilterTest, RequestDeny) {
  std::string path = GetModelPath();
  setup(path, 0, Source::Request, Action::Deny);

  Buffer::InstancePtr buf(new Buffer::OwnedImpl());
  EXPECT_CALL(decoded_callbacks_, addDecodedData(_, _))
      .WillRepeatedly(Invoke([&](Buffer::Instance& data, bool) { buf->add(data); }));
  EXPECT_CALL(decoded_callbacks_, decodingBuffer()).WillRepeatedly(Return(buf.get()));

  Buffer::OwnedImpl data("{\"text\": \"What a beautiful world!\"}");
  EXPECT_EQ(Http::FilterDataStatus::StopIterationNoBuffer, filter_->decodeData(data, true));
  buf->drain(buf->length());
  Buffer::OwnedImpl data2("{\"text\": \"Free money!\"}");
  EXPECT_EQ(Http::FilterDataStatus::Continue, filter_->decodeData(data2, true));
}

TEST_F(FilterTest, ResponseAllow) {
  std::string path = GetModelPath();
  setup(path, 0, Source::Response, Action::Allow);

  Buffer::InstancePtr buf(new Buffer::OwnedImpl());
  EXPECT_CALL(encoded_callbacks_, addEncodedData(_, _))
      .WillRepeatedly(Invoke([&](Buffer::Instance& data, bool) { buf->add(data); }));
  EXPECT_CALL(encoded_callbacks_, encodingBuffer()).WillRepeatedly(Return(buf.get()));

  Buffer::OwnedImpl data("{\"text\": \"What a beautiful world!\"}");
  EXPECT_EQ(Http::FilterDataStatus::Continue, filter_->encodeData(data, true));
  buf->drain(buf->length());
  Buffer::OwnedImpl data2("{\"text\": \"Free money!\"}");
  EXPECT_EQ(Http::FilterDataStatus::StopIterationNoBuffer, filter_->encodeData(data2, true));
}

TEST_F(FilterTest, ResponseDeny) {
  std::string path = GetModelPath();
  setup(path, 0, Source::Response, Action::Deny);

  Buffer::InstancePtr buf(new Buffer::OwnedImpl());
  EXPECT_CALL(encoded_callbacks_, addEncodedData(_, _))
      .WillRepeatedly(Invoke([&](Buffer::Instance& data, bool) { buf->add(data); }));
  EXPECT_CALL(encoded_callbacks_, encodingBuffer()).WillRepeatedly(Return(buf.get()));

  Buffer::OwnedImpl data("{\"text\": \"What a beautiful world!\"}");
  EXPECT_EQ(Http::FilterDataStatus::StopIterationNoBuffer, filter_->encodeData(data, true));
  buf->drain(buf->length());
  Buffer::OwnedImpl data2("{\"text\": \"Free money!\"}");
  EXPECT_EQ(Http::FilterDataStatus::Continue, filter_->encodeData(data2, true));
}

} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
