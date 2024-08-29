#include "source/extensions/filters/http/guardrails/filter.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

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

    bool matched = session_->classify(decoder_callbacks_->decodingBuffer()->toString());
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

    bool matched = session_->classify(encoder_callbacks_->encodingBuffer()->toString());
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
