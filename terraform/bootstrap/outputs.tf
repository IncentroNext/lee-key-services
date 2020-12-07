
output "project" {
  value = var.project
}

output "secrets_file" {
  value = var.secrets_file
}

output "proj_hash" {
  value = regex("https://.+-([a-z0-9]{10})-ew\\.a\\.run\\.app", google_cloud_run_service.dummy_service.status[0].url)[0]
}
