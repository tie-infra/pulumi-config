package main

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed config.yml
var configFile []byte

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	var zones []Zone
	if err := yaml.Unmarshal(configFile, &zones); err != nil {
		return fmt.Errorf("decode embedded config: %w", err)
	}

	for _, z := range zones {
		if err := setupZone(ctx, &z); err != nil {
			return err
		}
	}
	return nil
}
