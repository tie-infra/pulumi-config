package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Config struct {
	Zones []Zone
}

func parseConfig(ctx *pulumi.Context) (*Config, error) {
	var c Config

	conf := config.New(ctx, "")
	if err := conf.GetObject("config", &c); err != nil {
		return nil, err
	}

	return &c, nil
}
