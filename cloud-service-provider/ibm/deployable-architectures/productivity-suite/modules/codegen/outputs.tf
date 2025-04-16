output "release_status" {
  description = "Status of the Codegen Helm release"
  value       = helm_release.codegen.status
}

output "namespace" {
  description = "Namespace where Codegen is deployed"
  value       = helm_release.codegen.namespace
}

output "chart_version" {
  description = "Deployed chart version"
  value       = helm_release.codegen.version
}

output "release_name" {
  description = "Name of the Helm release"
  value       = helm_release.codegen.name
}