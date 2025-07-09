output "org_name" {
  value       = module.app.org_name
  description = "Used by the terraform-cleanup step."
}

output "space_name" {
  value       = module.app.space_name
  description = "Used by the terraform-cleanup step."
}

output "app_name" {
  value       = module.app.app_name
  description = "Used by the terraform-cleanup step."
}
