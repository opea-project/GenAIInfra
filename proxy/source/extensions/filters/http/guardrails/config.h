#pragma once

#include "source/extensions/filters/http/common/factory_base.h"
#include "source/extensions/filters/http/guardrails/config.pb.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Guardrails {

class FilterConfigFactory : public Server::Configuration::NamedHttpFilterConfigFactory {
public:
  absl::StatusOr<Http::FilterFactoryCb>
  createFilterFactoryFromProto(const Protobuf::Message& config, const std::string& stat_prefix,
                               Server::Configuration::FactoryContext& context) override;
  ProtobufTypes::MessagePtr createEmptyConfigProto() override;
  std::string name() const override { return "envoy.filters.http.guardrails"; }

private:
  Http::FilterFactoryCb
  createFilterFactory(const envoy::extensions::filters::http::guardrails::v3::Guardrails& config_pb,
                      Server::Configuration::FactoryContext& context);
};

} // namespace Guardrails
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
