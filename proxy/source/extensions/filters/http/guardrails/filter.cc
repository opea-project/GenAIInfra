#include "source/extensions/filters/http/guardrails/filter.h"

#include "source/common/json/json_loader.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

namespace {
std::string getText(const std::string& buffer) {
  absl::StatusOr<Json::ObjectSharedPtr> loader = Json::Factory::loadFromStringNoThrow(buffer);
  if (!loader.ok()) {
    return "";
  }

  absl::StatusOr<Envoy::Json::ValueType> value_or_error;
  value_or_error = loader.value()->getValue("text");
  if (!value_or_error.ok() || !absl::holds_alternative<std::string>(value_or_error.value())) {
    return "";
  }

  return absl::get<std::string>(value_or_error.value());
}
} // namespace

void Filter::onDestroy() {}

Http::FilterHeadersStatus Filter::decodeHeaders(Http::RequestHeaderMap&, bool end_stream) {
  if (config_->source() == Source::Response || end_stream) {
    return Http::FilterHeadersStatus::Continue;
  }
  return Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::decodeData(Buffer::Instance& data, bool end_stream) {
  if (config_->source() == Source::Response) {
    return Http::FilterDataStatus::Continue;
  }

  if (end_stream) {
    decoder_callbacks_->addDecodedData(data, true);

    std::string content = decoder_callbacks_->decodingBuffer()->toString();
    bool matched = session_->classify(getText(content));
    if ((config_->action() == Action::Allow && matched) ||
        (config_->action() == Action::Deny && !matched)) {
      return Http::FilterDataStatus::Continue;
    }

    decoder_callbacks_->sendLocalReply(Http::Code::Forbidden, "Access denied", nullptr,
                                       absl::nullopt, "");
    return Http::FilterDataStatus::StopIterationNoBuffer;
  }

  return Http::FilterDataStatus::StopIterationAndBuffer;
}

Http::FilterTrailersStatus Filter::decodeTrailers(Http::RequestTrailerMap&) {
  return Http::FilterTrailersStatus::Continue;
}

void Filter::setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) {
  decoder_callbacks_ = &callbacks;
}

Http::Filter1xxHeadersStatus Filter::encode1xxHeaders(Http::ResponseHeaderMap&) {
  return Http::Filter1xxHeadersStatus::Continue;
}

Http::FilterHeadersStatus Filter::encodeHeaders(Http::ResponseHeaderMap&, bool end_stream) {
  if (config_->source() == Source::Request || end_stream) {
    return Http::FilterHeadersStatus::Continue;
  }
  return Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::encodeData(Buffer::Instance& data, bool end_stream) {
  if (config_->source() == Source::Request) {
    return Http::FilterDataStatus::Continue;
  }

  if (end_stream) {
    encoder_callbacks_->addEncodedData(data, true);

    std::string content = encoder_callbacks_->encodingBuffer()->toString();
    bool matched = session_->classify(getText(content));
    if ((config_->action() == Action::Allow && matched) ||
        (config_->action() == Action::Deny && !matched)) {
      return Http::FilterDataStatus::Continue;
    }

    encoder_callbacks_->sendLocalReply(Http::Code::InternalServerError, "Access denied", nullptr,
                                       absl::nullopt, "");
    return Http::FilterDataStatus::StopIterationNoBuffer;
  }

  return Http::FilterDataStatus::StopIterationAndBuffer;
}

Http::FilterTrailersStatus Filter::encodeTrailers(Http::ResponseTrailerMap&) {
  return Http::FilterTrailersStatus::Continue;
}

Http::FilterMetadataStatus Filter::encodeMetadata(Http::MetadataMap&) {
  return Http::FilterMetadataStatus::Continue;
}

void Filter::setEncoderFilterCallbacks(Http::StreamEncoderFilterCallbacks& callbacks) {
  encoder_callbacks_ = &callbacks;
}

} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
