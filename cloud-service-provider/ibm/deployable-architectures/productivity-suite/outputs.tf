output "chatqna_release_status" {
  description = "Status of the ChatQnA Helm release"
  value       = module.chatqna.release_status
}

output "codegen_release_status" {
  description = "Status of the Codegen Helm release"
  value       = module.codegen.release_status
}

output "docsum_release_status" {
  description = "Status of the Docsum Helm release"
  value       = module.docsum.release_status
}

output "chathistory_release_status" {
  description = "Status of the Chathistory Helm release"
  value       = module.chathistory-usvc.release_status
}

output "prompt_release_status" {
  description = "Status of the Prompt Helm release"
  value       = module.prompt-usvc.release_status
}

output "nginx_endpoint" {
  description = "The endpoint for the nginx central gateway"
  value       = "Once deployed, run: kubectl get svc -n nginx nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}'"
}