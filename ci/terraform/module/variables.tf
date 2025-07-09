variable "cloud_gov_environment" {
  type        = string
  description = "Like development, staging, or production."
}

# CF Application Configuration

variable "org_name" {
  type        = string
  description = "The name of the Cloud Foundry organization in which the broker will be deployed."
}

variable "space_name" {
  type        = string
  description = "The name of the Cloud Foundry space in which the broker will be deployed."
}

variable "instances" {
  description = "Number of instances of the CSB app to run."
  type        = number
}

variable "path" {
  description = "Path to the source for the app to be pushed"
}

variable "base_domain" {
  type        = string
  description = "The domain under which the service's route will be created. For example, 'fr.cloud.gov'."
}
