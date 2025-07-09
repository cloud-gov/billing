variable "stack_name" {
  type        = string
  description = "One of development, staging, production."
}

variable "aws_region_govcloud" {
  type        = string
  description = "The AWS region in GovCloud in which to deploy the billing service."
}

variable "org_name" {
  type = string
}

variable "space_name" {
  type = string
}

variable "base_domain" {
  type = string
}

variable "path" {
  type = string
}
