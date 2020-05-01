/*
 * Terraform configuration for deploying necessary elements
 */

variable "region" {
  default = "us-west-2"
}
variable "s3_endpoint" {
  default = ""
}
variable "s3_force_path_style" {
  default = false
}
variable "access_key" {
  default = ""
}
variable "secret_key" {
  default = ""
}
variable "skip_meta_api_check" {
  default = false
}
variable "skip_credentials_validation" {
  default = false
}
variable "skip_requesting_account_id" {
  default = false
}
provider "aws" {
  profile = "default"
  access_key = var.access_key
  secret_key = var.secret_key
  region  = var.region
  s3_force_path_style = var.s3_force_path_style
  skip_credentials_validation = var.skip_credentials_validation
  skip_metadata_api_check     = var.skip_meta_api_check
  skip_requesting_account_id  = var.skip_requesting_account_id
  endpoints {
    s3 = var.s3_endpoint
  }
}

resource "aws_s3_bucket" "audio-bucket" {
  bucket = "converted-audio-${uuid()}"
  acl    = "private"
  lifecycle_rule {
    enabled = true
    expiration {
      days = 1
    }
  }
}

output "s3_bucket_id" {
  value = aws_s3_bucket.audio-bucket.id
  description = "The ID of the bucket that grpc-audio-converter saves converted files to"
}

