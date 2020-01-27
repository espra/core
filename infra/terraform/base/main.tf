// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

terraform {
  backend "remote" {
  }
}

provider "google" {
  version = "= 3.4.0"
}

data "google_compute_image" "cos" {
  family  = "cos-stable"
  project = "cos-cloud"
}

data "google_compute_network" "base" {
  name    = "default"
  project = google_project.base.project_id
}

resource "google_compute_instance_template" "base" {
  description          = "Template for creating Espra Website instances"
  instance_description = "Espra Website"
  machine_type         = "n1-standard-1"
  min_cpu_platform     = "Intel Cascade Lake"
  project              = google_project.base.project_id
  name_prefix          = "base-"
  region               = "us-central1"
  tags                 = ["base"]
  disk {
    auto_delete  = true
    boot         = true
    source_image = data.google_compute_image.cos.self_link
  }
  lifecycle {
    create_before_destroy = true
  }
  metadata = {
    gce-container-declaration = yamlencode({
      "image" : "${var.base_project_id}"
    })
  }
  network_interface {
    network = "default"
  }
  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }
  service_account {
    email  = ""
    scopes = []
  }
}

resource "google_compute_region_instance_group_manager" "base" {
  base_instance_name = "base"
  name               = "base"
  project            = google_project.base.project_id
  region             = "us-central1"
  target_size        = 1
  version {
    instance_template = google_compute_instance_template.base.self_link
  }
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
