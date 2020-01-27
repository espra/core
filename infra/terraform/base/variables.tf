// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

variable "base_project_id" {
  type        = string
  description = "The ID of the Google Cloud project for Base"
}

variable "google_billing_account" {
  type        = string
  description = "The alphanumeric ID of the Google Cloud billing account"
}

variable "google_org_id" {
  type        = string
  description = "The numeric ID of the Google Cloud organization"
}
