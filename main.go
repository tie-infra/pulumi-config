package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	conf, err := parseConfig(ctx)
	if err != nil {
		return err
	}
	if err := setupZones(ctx, conf); err != nil {
		return err
	}
	return nil
}
