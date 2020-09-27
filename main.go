package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	localtest "github.com/takinaga-dev/terraform-provider-localtest/local"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: localtest.Provider})
}
