output "release_status" {
  description = "Status of the Chathistory Helm release"
  value       = helm_release.chathistory-usvc.status
}
