// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

variable "base_project_id" {
  type        = string
  description = "The ID of the Google Cloud base project"
}

variable "loadbalancer_logs_bucket" {
  type        = string
  description = "The ID of the Google Cloud Storage bucket for the website loadbalancer logs"
}
