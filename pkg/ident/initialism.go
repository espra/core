// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

package ident

import (
	"strings"
)

var mapping = map[string]string{}

// This list helps us satisfy the recommended naming style of variables in Go:
// https://github.com/golang/go/wiki/CodeReviewComments#initialisms
//
// The list is always going to be incomplete, so please add to it as we come
// across new initialisms.
var initialisms = []string{
	"ACK",
	"ACL",
	"ACLs",
	"AES",
	"ANSI",
	"API",
	"APIs",
	"ARP",
	"ASCII",
	"ASN1",
	"ATM",
	"BGP",
	"BIOS",
	"BLAKE",
	"BLAKE3",
	"BSS",
	"CA",
	"CIDR",
	"CLI",
	"CLUI",
	"CPU",
	"CPUs",
	"CRC",
	"CSRF",
	"CSS",
	"CSV",
	"DB",
	"DBs",
	"DHCP",
	"DNS",
	"DRM",
	"EOF",
	"EON",
	"FTP",
	"GRPC",
	"GUID",
	"GUIDs",
	"HCL",
	"HTML",
	"HTTP",
	"HTTPS",
	"IANA",
	"ICMP",
	"ID",
	"IDs",
	"IEEE",
	"IMAP",
	"IP",
	"IPs",
	"IRC",
	"ISO",
	"ISP",
	"JSON",
	"LAN",
	"LHS",
	"MAC",
	"MD5",
	"MTU",
	"NATO",
	"NIC",
	"NVRAM",
	"OSI",
	"PEM",
	"POP3",
	"QPS",
	"QUIC",
	"RAM",
	"RFC",
	"RFCs",
	"RHS",
	"RPC",
	"SFTP",
	"SHA",
	"SHA1",
	"SHA256",
	"SHA512",
	"SLA",
	"SMTP",
	"SQL",
	"SRAM",
	"SSH",
	"SSID",
	"SSL",
	"SYN",
	"TCP",
	"TLS",
	"TOML",
	"TPS",
	"TTL",
	"UDP",
	"UI",
	"UID",
	"UIDs",
	"URI",
	"URL",
	"USB",
	"UTF8",
	"UUID",
	"UUIDs",
	"VLAN",
	"VM",
	"VPN",
	"W3C",
	"WPA",
	"WiFi",
	"XML",
	"XMPP",
	"XON",
	"XSRF",
	"XSS",
	"YAML",
}

// AddInitialism adds the given identifier to the set of initialisms. The given
// identifier should be in the PascalCase form and have at most one lower-cased
// letter which must be at the very end.
func AddInitialism(ident string) {
	mapping[strings.ToUpper(ident)] = ident
}

func init() {
	for _, s := range initialisms {
		mapping[strings.ToUpper(s)] = s
	}
}
