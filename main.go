package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/labd/terraform-provider-contentstack/internal/provider"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/labd/contentstack",
		Debug:   debugMode,
	})
}
