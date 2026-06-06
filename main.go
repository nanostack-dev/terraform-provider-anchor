package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	providerpkg "github.com/nanostack-dev/terraform-provider-anchor/internal/provider"
)

// version is set at build time via -ldflags by GoReleaser.
var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), providerpkg.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/nanostack-dev/anchor",
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}
