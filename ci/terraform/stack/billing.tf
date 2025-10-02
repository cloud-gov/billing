locals {
  billing_route = "billing.${var.base_domain}"
}

data "cloudfoundry_org" "platform" {
  name = var.org_name
}

data "cloudfoundry_space" "services" {
  name = var.space_name
  org  = data.cloudfoundry_org.platform.id
}

data "archive_file" "app_zip" {
  type        = "zip"
  source_dir  = var.path
  output_path = "./billing.zip"
}

resource "cloudfoundry_app" "billing" {
  name       = "billing"
  org_name   = var.org_name
  space_name = var.space_name

  path             = data.archive_file.app_zip.output_path
  buildpacks       = ["go_buildpack"]
  source_code_hash = data.archive_file.app_zip.output_base64sha256

  command    = "billing"
  instances  = var.instances
  memory     = "128M"
  disk_quota = "2G"

  environment = merge(
    {
      "GO_LINKER_SYMBOL" = "main.BuildVersion"
      "GO_LINKER_VALUE"  = var.short_ref
      "GOVERSION"        = "1.24"
    },
    var.environment
  )

  routes = [{
    route = local.billing_route
  }]

  service_bindings = [{
    service_instance = cloudfoundry_service_instance.db.name
  }]
}

data "cloudfoundry_service_plan" "rds" {
  name                  = "small-psql"
  service_offering_name = "aws-rds"
}

resource "cloudfoundry_service_instance" "db" {
  name         = "billing-db"
  space        = data.cloudfoundry_space.services.id
  type         = "managed"
  service_plan = data.cloudfoundry_service_plan.rds.id
  parameters   = jsonencode({ "storage" : 20 })

  # Prevent deleting the database by accident.
  lifecycle {
    prevent_destroy = true
  }
}
