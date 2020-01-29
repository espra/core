// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

terraform {
  backend "remote" {
  }
}

provider "google" {
  project = var.base_project_id
  region  = "us-central1"
  version = "= 3.5.0"
}

provider "google-beta" {
  project = var.base_project_id
  region  = "us-central1"
  version = "= 3.5.0"
}

data "google_compute_image" "cos" {
  family  = "cos-stable"
  project = "cos-cloud"
}

data "google_compute_network" "default" {
  name = "default"
}

resource "google_compute_backend_service" "website" {
  provider                        = google-beta
  connection_draining_timeout_sec = 30
  custom_request_headers = [
    "X-City:{client_city}",
    "X-LatLong:{client_city_lat_long}",
    "X-Region:{client_region}",
  ]
  health_checks = [google_compute_health_check.website_loadbalancer.self_link]
  name          = "website"
  port_name     = "https"
  protocol      = "HTTPS"
  backend {
    group = google_compute_region_instance_group_manager.website.instance_group
  }
  log_config {
    enable      = true
    sample_rate = 1.0
  }
}

resource "google_compute_firewall" "load_balancer" {
  name    = "website-loadbalancer"
  network = data.google_compute_network.default.name
  source_ranges = [
    "35.191.0.0/16",
    "130.211.0.0/22",
  ]
  allow {
    protocol = "tcp"
    ports    = ["443"]
  }
}

resource "google_compute_global_address" "ipv4" {
  ip_version = "IPV4"
  name       = "website-ipv4"
}

resource "google_compute_global_address" "ipv6" {
  ip_version = "IPV6"
  name       = "website-ipv6"
}

resource "google_compute_global_forwarding_rule" "ipv4_http" {
  ip_address = google_compute_global_address.ipv4.address
  name       = "website-ipv4-http"
  port_range = "80"
  target     = google_compute_target_http_proxy.website.self_link
}

resource "google_compute_global_forwarding_rule" "ipv4_https" {
  ip_address = google_compute_global_address.ipv4.address
  name       = "website-ipv4-https"
  port_range = "443"
  target     = google_compute_target_https_proxy.website.self_link
}

resource "google_compute_global_forwarding_rule" "ipv6_http" {
  ip_address = google_compute_global_address.ipv6.address
  name       = "website-ipv6-http"
  port_range = "80"
  target     = google_compute_target_http_proxy.website.self_link
}

resource "google_compute_global_forwarding_rule" "ipv6_https" {
  ip_address = google_compute_global_address.ipv6.address
  name       = "website-ipv6-https"
  port_range = "443"
  target     = google_compute_target_https_proxy.website.self_link
}

resource "google_compute_health_check" "website_instance" {
  check_interval_sec  = 10
  name                = "website-instance"
  timeout_sec         = 10
  unhealthy_threshold = 3
  https_health_check {
    host         = "dappui.com"
    request_path = "/health"
    response     = "OK"
  }
}

resource "google_compute_health_check" "website_loadbalancer" {
  name = "website-loadbalancer"
  https_health_check {
    host         = "dappui.com"
    request_path = "/health"
    response     = "OK"
  }
}

resource "google_compute_instance_template" "website" {
  description          = "Template for creating Website instances"
  instance_description = "Website"
  machine_type         = "n1-standard-1"
  min_cpu_platform     = "Intel Skylake"
  name_prefix          = "website-"
  tags                 = ["website"]
  disk {
    boot         = true
    source_image = data.google_compute_image.cos.self_link
  }
  lifecycle {
    create_before_destroy = true
  }
  metadata = {
    gce-container-declaration = yamlencode({
      spec = {
        containers = [{
          image = "gcr.io/${var.base_project_id}/website:first"
          securityContext = {
            privileged = true
          }
        }]
        restartPolicy = "Always"
      }
    })
  }
  network_interface {
    network = data.google_compute_network.default.name
    access_config {
      network_tier = "PREMIUM"
    }
  }
  scheduling {
    on_host_maintenance = "MIGRATE"
  }
  service_account {
    email  = google_service_account.website.email
    scopes = ["storage-ro"]
  }
}

resource "google_compute_managed_ssl_certificate" "website" {
  provider = google-beta
  name     = "website"
  managed {
    domains = ["dappui.com", "www.dappui.com"]
  }
}

resource "google_compute_region_instance_group_manager" "website" {
  base_instance_name        = "website"
  distribution_policy_zones = ["us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"]
  name                      = "website"
  region                    = "us-central1"
  target_size               = 1
  auto_healing_policies {
    health_check      = google_compute_health_check.website_instance.self_link
    initial_delay_sec = 300
  }
  named_port {
    name = "https"
    port = 443
  }
  update_policy {
    instance_redistribution_type = "PROACTIVE"
    max_surge_fixed              = 4
    minimal_action               = "REPLACE"
    type                         = "PROACTIVE"
  }
  version {
    instance_template = google_compute_instance_template.website.self_link
  }
}

resource "google_compute_ssl_policy" "website" {
  min_tls_version = "TLS_1_2"
  name            = "website"
  profile         = "RESTRICTED"
}

resource "google_compute_target_http_proxy" "website" {
  name    = "website"
  url_map = google_compute_url_map.website.self_link
}

resource "google_compute_target_https_proxy" "website" {
  name             = "website"
  ssl_certificates = [google_compute_managed_ssl_certificate.website.self_link]
  ssl_policy       = google_compute_ssl_policy.website.self_link
  url_map          = google_compute_url_map.website.self_link
}

resource "google_compute_url_map" "website" {
  default_service = google_compute_backend_service.website.self_link
  name            = "website"
}

resource "google_logging_project_sink" "loadbalancer" {
  name                   = "website-loadbalancer"
  destination            = "storage.googleapis.com/${var.loadbalancer_logs_bucket}"
  filter                 = "resource.type=\"http_load_balancer\" AND (resource.labels.forwarding_rule_name=\"website-ipv4-http\" OR resource.labels.forwarding_rule_name=\"website-ipv4-https\" OR resource.labels.forwarding_rule_name=\"website-ipv6-http\" OR resource.labels.forwarding_rule_name=\"website-ipv6-https\")"
  unique_writer_identity = true
}

resource "google_project_iam_member" "datastore" {
  member = "serviceAccount:${google_service_account.website.email}"
  role   = "roles/datastore.user"
}

resource "google_service_account" "website" {
  account_id   = "website"
  display_name = "Website Service Account"
}

resource "google_storage_bucket" "loadbalancer_logs" {
  name = var.loadbalancer_logs_bucket
}

resource "google_storage_bucket_iam_member" "container_registry" {
  bucket = "artifacts.${var.base_project_id}.appspot.com"
  member = "serviceAccount:${google_service_account.website.email}"
  role   = "roles/storage.objectViewer"
}

resource "google_storage_bucket_iam_member" "loadbalancer_logs" {
  depends_on = [google_storage_bucket.loadbalancer_logs]
  bucket     = var.loadbalancer_logs_bucket
  member     = google_logging_project_sink.loadbalancer.writer_identity
  role       = "roles/storage.objectCreator"
}
