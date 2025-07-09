variable "stack_name" {
  type        = string
  description = "One of development, staging, production."
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
