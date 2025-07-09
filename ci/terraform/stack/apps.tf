module "app" {
  source = "../module"

  cloud_gov_environment = var.stack_name

  instances = 1

  org_name   = var.org_name
  space_name = var.space_name

  base_domain = var.base_domain
  path        = var.path
}
