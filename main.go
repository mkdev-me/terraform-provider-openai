package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/mkdev-me/terraform-provider-openai/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "2.0.0"

	// goreleaser can also tag the specific commit that release was built on.
	// commit  string = ""
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/mkdev-me/openai",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.NewFrameworkProvider(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
