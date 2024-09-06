/*
 * Copyright (C) 2024 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

#pragma once

#include <deque>

#include "envoy/http/filter.h"

#include "source/common/buffer/buffer_impl.h"
#include "source/extensions/common/inference/inference_runtime.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

enum class Source { Request, Response };

enum class Action { Allow, Deny };

class FilterConfig {
public:
  FilterConfig(Extensions::Common::Inference::InferenceRuntimePtr runtime, Source source,
               Action action)
      : runtime_(std::move(runtime)), source_(source), action_(action) {}

  Extensions::Common::Inference::InferenceSessionPtr createInferenceSession() const {
    return runtime_->createInferenceSession();
  }
  Source source() const { return source_; }
  Action action() const { return action_; }

protected:
  Extensions::Common::Inference::InferenceRuntimePtr runtime_;
  Source source_;
  Action action_;
};

using FilterConfigSharedPtr = std::shared_ptr<FilterConfig>;

class Filter : public Http::StreamFilter, Logger::Loggable<Logger::Id::filter> {
public:
  Filter(FilterConfigSharedPtr config)
      : config_(config), session_(config_->createInferenceSession()) {}

  // Http::StreamFilterBase
  void onDestroy() override;

  // Http::StreamDecoderFilter
  Http::FilterHeadersStatus decodeHeaders(Http::RequestHeaderMap& headers,
                                          bool end_stream) override;
  Http::FilterDataStatus decodeData(Buffer::Instance& data, bool end_stream) override;
  Http::FilterTrailersStatus decodeTrailers(Http::RequestTrailerMap& trailers) override;
  void setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) override;

  // Http::StreamEncoderFilter
  Http::Filter1xxHeadersStatus encode1xxHeaders(Http::ResponseHeaderMap& headers) override;
  Http::FilterHeadersStatus encodeHeaders(Http::ResponseHeaderMap& headers,
                                          bool end_stream) override;
  Http::FilterDataStatus encodeData(Buffer::Instance& data, bool end_stream) override;
  Http::FilterTrailersStatus encodeTrailers(Http::ResponseTrailerMap& trailers) override;
  Http::FilterMetadataStatus encodeMetadata(Http::MetadataMap& metadata_map) override;
  void setEncoderFilterCallbacks(Http::StreamEncoderFilterCallbacks& callbacks) override;

protected:
  FilterConfigSharedPtr config_;
  Extensions::Common::Inference::InferenceSessionPtr session_;
  Http::StreamDecoderFilterCallbacks* decoder_callbacks_{};
  Http::StreamEncoderFilterCallbacks* encoder_callbacks_{};
};

} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
