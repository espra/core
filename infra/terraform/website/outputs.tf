// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

output "ipv4_address" {
  description = "The IPv4 address for the Website"
  value       = google_compute_global_address.ipv4.address
}

output "ipv6_address" {
  description = "The IPv6 address for the Website"
  value       = google_compute_global_address.ipv6.address
}
