output "org_name" {
  value       = var.org_name
  description = "Used by the terraform-cleanup step."
}

output "space_name" {
  value       = var.space_name
  description = "Used by the terraform-cleanup step."
}

output "app_name" {
  value       = cloudfoundry_app.billing.name
  description = "Used by the terraform-cleanup step."
}
