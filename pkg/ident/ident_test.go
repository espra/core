// Public Domain (-) 2020-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package ident

import (
	"strings"
	"testing"
)

var spec = map[string]*definition{
	"HTTPSServer": {
		camel: "httpsServer",
		kebab: "https-server",
	},
	"I": {
		camel: "i",
		kebab: "i",
	},
	"IDSet": {
		camel: "idSet",
		kebab: "id-set",
	},
	"IDs": {
		camel: "ids",
		kebab: "ids",
	},
	"IDsMap": {
		camel: "idsMap",
		kebab: "ids-map",
	},
	"NetworkCIDR": {
		camel: "networkCIDR",
		kebab: "network-cidr",
	},
	"PCRTestKit": {
		camel: "pcrTestKit",
		kebab: "pcr-test-kit",
	},
	"PeerAPIOp": {
		camel: "peerAPIOp",
		kebab: "peer-api-op",
	},
	"PeerIDs": {
		camel: "peerIDs",
		kebab: "peer-ids",
	},
	"ServiceAPIKey": {
		camel: "serviceAPIKey",
		kebab: "service-api-key",
	},
	"ServiceKey": {
		camel: "serviceKey",
		kebab: "service-key",
	},
	"UserACLIDs": {
		camel: "userACLIDs",
		kebab: "user-acl-ids",
	},
	"Username": {
		camel: "username",
		kebab: "username",
	},
	"XMLHTTP": {
		camel: "xmlHTTP",
		kebab: "xml-http",
	},
	"XMLHTTPRequest": {
		camel: "xmlHTTPRequest",
		kebab: "xml-http-request",
	},
}

var tests = []testcase{
	{"https-server", spec["HTTPSServer"]},
	{"https-server-", spec["HTTPSServer"]},
	{"-https-server", spec["HTTPSServer"]},
	{"--https-server-", spec["HTTPSServer"]},
	{"ids", spec["IDs"]},
	{"ids-", spec["IDs"]},
	{"-ids", spec["IDs"]},
	{"--ids-", spec["IDs"]},
	{"ids-map", spec["IDsMap"]},
	{"ids-map-", spec["IDsMap"]},
	{"-ids-map", spec["IDsMap"]},
	{"--ids-map-", spec["IDsMap"]},
	{"network-cidr", spec["NetworkCIDR"]},
	{"network-cidr-", spec["NetworkCIDR"]},
	{"-network-cidr", spec["NetworkCIDR"]},
	{"--network-cidr-", spec["NetworkCIDR"]},
	{"peer-api-op", spec["PeerAPIOp"]},
	{"peer-api-op-", spec["PeerAPIOp"]},
	{"-peer-api-op", spec["PeerAPIOp"]},
	{"--peer-api-op-", spec["PeerAPIOp"]},
	{"peer-ids", spec["PeerIDs"]},
	{"peer-ids-", spec["PeerIDs"]},
	{"-peer-ids", spec["PeerIDs"]},
	{"--peer-ids-", spec["PeerIDs"]},
	{"service-api-key", spec["ServiceAPIKey"]},
	{"service-api-key-", spec["ServiceAPIKey"]},
	{"-service-api-key", spec["ServiceAPIKey"]},
	{"--service-api-key-", spec["ServiceAPIKey"]},
	{"service-key", spec["ServiceKey"]},
	{"service-key-", spec["ServiceKey"]},
	{"-service-key", spec["ServiceKey"]},
	{"--service-key-", spec["ServiceKey"]},
	{"user-acl-ids", spec["UserACLIDs"]},
	{"user-acl-ids-", spec["UserACLIDs"]},
	{"-user-acl-ids", spec["UserACLIDs"]},
	{"--user-acl-ids-", spec["UserACLIDs"]},
	{"username", spec["Username"]},
	{"username-", spec["Username"]},
	{"-username", spec["Username"]},
	{"--username-", spec["Username"]},
	{"xml-http", spec["XMLHTTP"]},
	{"xml-http-", spec["XMLHTTP"]},
	{"-xml-http", spec["XMLHTTP"]},
	{"--xml-http-", spec["XMLHTTP"]},
	{"xml-http-request", spec["XMLHTTPRequest"]},
	{"xml-http-request-", spec["XMLHTTPRequest"]},
	{"-xml-http-request", spec["XMLHTTPRequest"]},
	{"--xml-http-request-", spec["XMLHTTPRequest"]},
}

type definition struct {
	camel     string
	kebab     string
	pascal    string
	screaming string
	snake     string
}

type testcase struct {
	ident string
	want  *definition
}

func TestCamel(t *testing.T) {
	for _, tt := range spec {
		testConversion(t, "Camel", FromCamel, tt.camel, tt)
	}
}

func TestKebab(t *testing.T) {
	for _, tt := range tests {
		testConversion(t, "Kebab", FromKebab, tt.ident, tt.want)
	}
}

func TestPascal(t *testing.T) {
	MustPascal := func(ident string) Parts {
		parts, err := FromPascal(ident)
		if err != nil {
			t.Fatalf("FromPascal(%q) returned an unexpected error: %s", ident, err)
		}
		return parts
	}
	_, err := FromPascal("invalid")
	if err == nil {
		t.Errorf("FromPascal(%q) failed to return an error", "invalid")
	}
	for _, tt := range spec {
		testConversion(t, "Pascal", MustPascal, tt.pascal, tt)
	}
}

func TestScreamingSnake(t *testing.T) {
	for _, tt := range tests {
		ident := strings.ToUpper(strings.ReplaceAll(tt.ident, "-", "_"))
		testConversion(t, "ScreamingSnake", FromScreamingSnake, ident, tt.want)
	}
}

func TestSnake(t *testing.T) {
	for _, tt := range tests {
		ident := strings.ReplaceAll(tt.ident, "-", "_")
		testConversion(t, "Snake", FromSnake, ident, tt.want)
	}
}

func TestString(t *testing.T) {
	ident := "HTTPAPIs"
	parts, _ := FromPascal(ident)
	got := parts.String()
	want := "HTTP,APIs"
	if got != want {
		t.Errorf("FromPascal(%q).String() = %q: want %q", ident, got, want)
	}
}

func testConversion(t *testing.T, typ string, conv func(string) Parts, ident string, want *definition) {
	id := conv(ident)
	got := id.ToCamel()
	if got != want.camel {
		t.Errorf("From%s(%q).ToCamel() = %q: want %q", typ, ident, got, want.camel)
	}
	got = id.ToKebab()
	if got != want.kebab {
		t.Errorf("From%s(%q).ToKebab() = %q: want %q", typ, ident, got, want.kebab)
	}
	got = id.ToPascal()
	if got != want.pascal {
		t.Errorf("From%s(%q).ToPascal() = %q: want %q", typ, ident, got, want.pascal)
	}
	got = id.ToScreamingSnake()
	if got != want.screaming {
		t.Errorf("From%s(%q).ToScreamingSnake() = %q: want %q", typ, ident, got, want.screaming)
	}
	got = id.ToSnake()
	if got != want.snake {
		t.Errorf("From%s(%q).ToSnake() = %q: want %q", typ, ident, got, want.snake)
	}
}

func init() {
	for pascal, definition := range spec {
		definition.pascal = pascal
		definition.snake = strings.ReplaceAll(definition.kebab, "-", "_")
		definition.screaming = strings.ToUpper(definition.snake)
	}
	AddInitialism("PCR")
}
