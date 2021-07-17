package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	var zones []Zone

	conf := config.New(ctx, "")
	if err := conf.GetObject("zones", &zones); err != nil {
		return err
	}

	for _, z := range zones {
		if err := setupZone(ctx, &z); err != nil {
			return err
		}
	}
	return nil
}
