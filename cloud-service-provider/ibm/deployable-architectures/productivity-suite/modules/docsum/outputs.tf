output "release_status" {
  description = "Status of the Docsum Helm release"
  value       = helm_release.docsum.status
}

output "namespace" {
  description = "Namespace where Docsum is deployed"
  value       = helm_release.docsum.namespace
}

output "chart_version" {
  description = "Deployed chart version"
  value       = helm_release.docsum.version
}

output "release_name" {
  description = "Name of the Helm release"
  value       = helm_release.docsum.name
}