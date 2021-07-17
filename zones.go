package main

import (
	"strings"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Zone struct {
	ID string

	Domains  []Domain
	Hosts    []Host
	Peers    []Peer
	Services []Service
}

type Domain struct {
	ID   string
	Name string
}

type Host struct {
	Name      string
	Addresses []HostAddress
}

type HostAddress struct {
	ID    string
	Value string
}

type Peer struct {
	Name string
	Host string
}

type Service struct {
	ID      string
	Service string
	Proto   string
	Name    string
	Prio    int
	Weight  int
	Host    string
	Port    int
}

func setupZone(ctx *pulumi.Context, z *Zone) error {
	for _, d := range z.Domains {
		if err := setupZoneDomain(ctx, z, &d); err != nil {
			return err
		}
	}
	return nil
}

func setupZoneDomain(ctx *pulumi.Context, z *Zone, d *Domain) error {
	resourceName := joinDash(z.ID, d.ID)

	zone, err := cloudflare.NewZone(ctx, resourceName, &cloudflare.ZoneArgs{
		Zone: pulumi.String(d.Name),
	})
	if err != nil {
		return err
	}
	ctx.Export(resourceName, zone.ID())

	if _, err := cloudflare.NewZoneSettingsOverride(ctx, resourceName, &cloudflare.ZoneSettingsOverrideArgs{
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

	if err := setupHosts(ctx, zone, z, d); err != nil {
		return err
	}
	if err := setupPeers(ctx, zone, z, d); err != nil {
		return err
	}
	if err := setupServices(ctx, zone, z, d); err != nil {
		return err
	}
	return nil
}

func setupHosts(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain) error {
	for _, host := range z.Hosts {
		if err := setupHost(ctx, zone, z, d, &host); err != nil {
			return err
		}
	}
	return nil
}

func setupHost(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain, host *Host) error {
	for _, addr := range host.Addresses {
		if err := setupHostAddress(ctx, zone, z, d, host, &addr); err != nil {
			return err
		}
	}
	return nil
}

func setupHostAddress(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain, host *Host, addr *HostAddress) error {
	resourceName := joinDash(z.ID, d.ID, host.Name, addr.ID)

	typ := "AAAA"
	if !strings.Contains(addr.Value, ":") {
		typ = "A"
	}

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String(host.Name),
		Value:  pulumi.String(addr.Value),
		Type:   pulumi.String(typ),
	})
	return err
}

func setupPeers(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain) error {
	for _, peer := range z.Peers {
		if err := setupPeer(ctx, zone, z, d, &peer); err != nil {
			return err
		}
	}
	return nil
}

func setupPeer(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain, peer *Peer) error {
	resourceName := joinDash(z.ID, d.ID, peer.Name)

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String(peer.Name),
		Value:  pulumi.String(peer.Host),
		Type:   pulumi.String("CNAME"),
	})
	return err
}

func setupServices(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain) error {
	for _, srv := range z.Services {
		if err := setupService(ctx, zone, z, d, &srv); err != nil {
			return err
		}
	}
	return nil
}

func setupService(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain, srv *Service) error {
	resourceName := joinDash(z.ID, d.ID, srv.Name, srv.ID)

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Data: &cloudflare.RecordDataArgs{
			Service:  pulumi.String(srv.Service),
			Proto:    pulumi.String(srv.Proto),
			Priority: pulumi.Int(srv.Prio),
			Weight:   pulumi.Int(srv.Weight),
			Port:     pulumi.Int(srv.Port),
			Target:   pulumi.String(srv.Host),
		},
		Name: pulumi.String(srv.Name),
		Type: pulumi.String("SRV"),
	})
	return err
}

func joinDash(elems ...string) string {
	return join("-", elems...)
}

func join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}
