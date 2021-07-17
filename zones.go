package main

import (
	"strings"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Zone struct {
	Domain string

	Addresses map[string][]ZoneAddress
	Aliases   map[string]string
	Services  map[string][]ZoneService
}

type ZoneAddress struct {
	ID string

	Address string
}

type ZoneService struct {
	ID string

	Service string
	Proto   string
	Prio    int
	Weight  int
	Host    string
	Port    int
}

func setupZones(ctx *pulumi.Context, conf *Config) error {
	for _, z := range conf.Zones {
		if err := setupZone(ctx, &z); err != nil {
			return err
		}
	}
	return nil
}

func setupZone(ctx *pulumi.Context, z *Zone) error {
	zone, err := cloudflare.NewZone(ctx, z.Domain, &cloudflare.ZoneArgs{
		Zone: pulumi.String(z.Domain),
	})
	if err != nil {
		return err
	}

	if _, err := cloudflare.NewZoneSettingsOverride(ctx, z.Domain, &cloudflare.ZoneSettingsOverrideArgs{
		ZoneId: zone.ID(),
		Settings: &cloudflare.ZoneSettingsOverrideSettingsArgs{
			Ssl:           pulumi.String("strict"),
			MinTlsVersion: pulumi.String("1.2"),
			ZeroRtt:       pulumi.String("on"),
			UniversalSsl:  pulumi.String("on"),
		},
	}); err != nil {
		return err
	}

	if err := setupAddresses(ctx, zone, z); err != nil {
		return err
	}
	if err := setupAliases(ctx, zone, z); err != nil {
		return err
	}
	if err := setupServices(ctx, zone, z); err != nil {
		return err
	}
	return nil
}

func setupAddresses(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone) error {
	for name, addresses := range z.Addresses {
		for _, record := range addresses {
			if err := setupAddress(ctx, zone, z, name, &record); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupAddress(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, name string, record *ZoneAddress) error {
	resourceName := name + "." + z.Domain + "/" + record.ID

	addr := record.Address

	typ := "AAAA"
	if !strings.Contains(addr, ":") {
		typ = "A"
	}

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String(name),
		Value:  pulumi.String(addr),
		Type:   pulumi.String(typ),
	})
	return err
}

func setupAliases(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone) error {
	for name, host := range z.Aliases {
		if err := setupAlias(ctx, zone, z, name, host); err != nil {
			return err
		}
	}
	return nil
}

func setupAlias(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, name, host string) error {
	resourceName := name + "." + z.Domain

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String(name),
		Value:  pulumi.String(host),
		Type:   pulumi.String("CNAME"),
	})
	return err
}

func setupServices(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone) error {
	for name, services := range z.Services {
		for _, record := range services {
			if err := setupService(ctx, zone, z, name, &record); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupService(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, name string, record *ZoneService) error {
	resourceName := name + "." + z.Domain + "/" + record.ID

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Data: &cloudflare.RecordDataArgs{
			Service:  pulumi.String(record.Service),
			Proto:    pulumi.String(record.Proto),
			Priority: pulumi.Int(record.Prio),
			Weight:   pulumi.Int(record.Weight),
			Port:     pulumi.Int(record.Port),
			Target:   pulumi.String(record.Host),
		},
		Name: pulumi.String(name),
		Type: pulumi.String("SRV"),
	})
	return err
}
