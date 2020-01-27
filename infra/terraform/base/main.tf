// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

terraform {
  backend "remote" {
  }
}

provider "google" {
  version = "= 3.4.0"
}

resource "google_folder" "core" {
  display_name = "Core"
  parent       = "organizations/${var.google_org_id}"
}

resource "google_project" "base" {
  billing_account = var.google_billing_account
  folder_id       = google_folder.core.name
  name            = "DappUI Base"
  project_id      = var.base_project_id
}
