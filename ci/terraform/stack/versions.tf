terraform {
  required_version = "< 2.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "< 7.0.0"
    }
    cloudfoundry = {
      source  = "cloudfoundry/cloudfoundry"
      version = "< 2.0"
    }
  }
}
