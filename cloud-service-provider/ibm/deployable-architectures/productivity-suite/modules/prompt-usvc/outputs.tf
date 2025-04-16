output "release_status" {
  description = "Status of the Prompt Helm release"
  value       = helm_release.prompt-usvc.status
}
