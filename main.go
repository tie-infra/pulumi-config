package main

import (
	"strings"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	conf := config.New(ctx, "")
	domain := conf.Require("domain")
	return setupDNS(ctx, domain)
}

func setupDNS(ctx *pulumi.Context, domain string) error {
	zone, err := cloudflare.NewZone(ctx, domain, &cloudflare.ZoneArgs{
		Zone: pulumi.String(domain),
	})
	if err != nil {
		return err
	}

	if _, err := cloudflare.NewZoneSettingsOverride(ctx, domain, &cloudflare.ZoneSettingsOverrideArgs{
		ZoneId: zone.ID(),
		Settings: &cloudflare.ZoneSettingsOverrideSettingsArgs{
			Ssl:           pulumi.StringPtr("strict"),
			MinTlsVersion: pulumi.StringPtr("1.2"),
			ZeroRtt:       pulumi.StringPtr("on"),
		},
	}); err != nil {
		return err
	}

	if err := setupHosts(ctx, zone); err != nil {
		return err
	}
	if err := setupWeb(ctx, zone); err != nil {
		return err
	}
	if err := setupMinecraft(ctx, zone); err != nil {
		return err
	}
	return nil
}

func setupHosts(ctx *pulumi.Context, zone *cloudflare.Zone) error {
	const (
		prefix    = "2a02:2168:8fec:f600:"
		prefixISP = "2a02:2168:a0f:a2a3:"
	)

	records := []struct {
		ID    string
		Name  string
		Value string
	}{
		{"47a44abc", "roku", prefix + ":1"},
		{"cb50f141", "roku", prefixISP + ":3"},
		{"93370deb", "roku", "95.84.246.62"},

		{"bcea9c81", "ubernet", prefixISP + ":4"},
		{"90057569", "ubernet", "188.32.206.130"},

		{"77ae1019", "madara", prefix + ":942"},

		{"212e2ced", "saitama", prefix + ":39f"},
		{"e89a326c", "saitama", "188.255.3.141"},

		{"f7bf04d2", "tatsuya", prefixISP + ":2"},
		{"93cba5bd", "tatsuya", "37.110.66.21"},
	}

	for _, r := range records {
		typ := "AAAA"
		if !strings.Contains(r.Value, ":") {
			typ = "A"
		}
		_, err := cloudflare.NewRecord(ctx, r.ID, &cloudflare.RecordArgs{
			ZoneId: zone.ID(),
			Name:   pulumi.String(r.Name),
			Value:  pulumi.String(r.Value),
			Type:   pulumi.String(typ),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func setupWeb(ctx *pulumi.Context, zone *cloudflare.Zone) error {
	_, err := cloudflare.NewRecord(ctx, "7157850e", &cloudflare.RecordArgs{
		ZoneId:  zone.ID(),
		Name:    pulumi.String("brim"),
		Value:   pulumi.String("brim.ml"),
		Type:    pulumi.String("CNAME"),
		Proxied: pulumi.BoolPtr(true),
	})
	return err
}

func setupMinecraft(ctx *pulumi.Context, zone *cloudflare.Zone) error {
	_, err := cloudflare.NewRecord(ctx, "1490d796", &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String("mc"),
		Type:   pulumi.String("SRV"),
		Data: &cloudflare.RecordDataArgs{
			Service:  pulumi.String("_minecraft"),
			Proto:    pulumi.String("_tcp"),
			Name:     pulumi.String("mc"),
			Priority: pulumi.Int(0),
			Weight:   pulumi.Int(0),
			Port:     pulumi.Int(25565),
			Target:   pulumi.String("brim.ml"),
		},
	})
	return err
}
