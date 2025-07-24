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
  description = "Number of instances of the app to run."
  type        = number
  default     = 1
}

variable "path" {
  description = "Path to the source for the app to be pushed."
}

variable "base_domain" {
  type        = string
  description = "The domain under which the service's route will be created. For example, 'fr.cloud.gov'."
}

variable "environment" {
  description = "Environment variables to set on the app."
  type = object({
    CF_API_URL       = string
    CF_CLIENT_ID     = string
    CF_CLIENT_SECRET = string
  })
}
