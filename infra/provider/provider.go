// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package main

import (
	"fmt"
	"os"
	"runtime"

	"dappui.com/infra/provider/container"
	"dappui.com/infra/provider/sourcehash"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const version = "0.0.1"

func main() {
	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "osarch":
			fmt.Println(runtime.GOOS + "_" + runtime.GOARCH)
		case "version":
			fmt.Println(version)
		}
		os.Exit(0)
	}
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return &schema.Provider{
				DataSourcesMap: map[string]*schema.Resource{
					"core_sourcehash": sourcehash.Resource(),
				},
				ResourcesMap: map[string]*schema.Resource{
					"core_container": container.Resource(),
				},
			}
		},
	})
}
