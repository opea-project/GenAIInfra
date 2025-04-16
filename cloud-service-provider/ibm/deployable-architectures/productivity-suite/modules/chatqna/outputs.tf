output "release_status" {
  description = "Status of the ChatQNA Helm release"
  value       = helm_release.chatqna.status
}

output "namespace" {
  description = "Namespace where ChatQNA is deployed"
  value       = helm_release.chatqna.namespace
}

output "chart_version" {
  description = "Deployed chart version"
  value       = helm_release.chatqna.version
}

output "release_name" {
  description = "Name of the Helm release"
  value       = helm_release.chatqna.name
}