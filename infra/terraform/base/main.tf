// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

terraform {
  backend "remote" {
  }
}

provider "google" {
  project = var.base_project_id
  version = "= 3.5.0"
}

resource "google_folder" "core" {
  display_name = "Core"
  parent       = "organizations/${var.google_org_id}"
}

resource "google_project" "base" {
  billing_account = var.google_billing_account
  folder_id       = google_folder.core.name
  name            = "DappUI"
  project_id      = var.base_project_id
}

resource "google_project_service" "base_compute" {
  service = "compute.googleapis.com"
}

resource "google_project_service" "base_container_registry" {
  service = "containerregistry.googleapis.com"
}

resource "google_project_service" "base_datastore" {
  service = "datastore.googleapis.com"
}

resource "google_project_service" "base_firestore" {
  service = "firestore.googleapis.com"
}

resource "google_project_service" "base_storage" {
  service = "storage-component.googleapis.com"
}

resource "google_project_service" "base_storage_api" {
  service = "storage-api.googleapis.com"
}
