/*
 * Copyright (C) 2024 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

#include "source/extensions/filters/http/guardrails/config.h"

#include "source/extensions/common/inference/utility.h"
#include "source/extensions/filters/http/guardrails/filter.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

absl::StatusOr<Http::FilterFactoryCb>
FilterConfigFactory::createFilterFactoryFromProto(const Protobuf::Message& config,
                                                  const std::string&,
                                                  Server::Configuration::FactoryContext& context) {
  return createFilterFactory(
      dynamic_cast<const envoy::extensions::filters::http::guardrails::v3::Guardrails&>(config),
      context);
}

ProtobufTypes::MessagePtr FilterConfigFactory::createEmptyConfigProto() {
  return ProtobufTypes::MessagePtr{
      new envoy::extensions::filters::http::guardrails::v3::Guardrails};
}

Http::FilterFactoryCb FilterConfigFactory::createFilterFactory(
    const envoy::extensions::filters::http::guardrails::v3::Guardrails& proto_config,
    Server::Configuration::FactoryContext& context) {
  Extensions::Common::Inference::InferenceRuntimePtr runtime =
      Extensions::Common::Inference::createInferenceRuntime(proto_config.model_path(),
                                                            proto_config.threshold(), context);
  Source source;
  switch (proto_config.source()) {
  case envoy::extensions::filters::http::guardrails::v3::Guardrails::REQUEST:
    source = Source::Request;
    break;
  case envoy::extensions::filters::http::guardrails::v3::Guardrails::RESPONSE:
    source = Source::Response;
    break;
  default:
    PANIC_DUE_TO_CORRUPT_ENUM;
  }
  Action action;
  switch (proto_config.action()) {
  case envoy::extensions::filters::http::guardrails::v3::Guardrails::ALLOW:
    action = Action::Allow;
    break;
  case envoy::extensions::filters::http::guardrails::v3::Guardrails::DENY:
    action = Action::Deny;
    break;
  default:
    PANIC_DUE_TO_CORRUPT_ENUM;
  }

  FilterConfigSharedPtr filter_config{
      std::make_shared<FilterConfig>(std::move(runtime), source, action)};
  return [filter_config](Http::FilterChainFactoryCallbacks& callbacks) -> void {
    callbacks.addStreamFilter(std::make_unique<Filter>(filter_config));
  };
}

REGISTER_FACTORY(FilterConfigFactory, Server::Configuration::NamedHttpFilterConfigFactory);

} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
