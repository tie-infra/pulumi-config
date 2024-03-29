package main

import (
	"net"
	"strings"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Zone defines a group of domains with records.
type Zone struct {
	// ID is a unique identifier for the zone. It is used as part of the
	// Pulumi resource ID and is expected to be human-readable and
	// descriptive enough in the configuration file.
	ID string

	// Domains
	Domains  []Domain
	Hosts    []Host
	Aliases  []Alias
	Services []Service
}

// Domain defines a managed DNS domain.
type Domain struct {
	// ID is a unique identifier for the domain. It is used as part of the
	// Pulumi resource ID and is expected to be human-readable and
	// descriptive enough in the configuration file.
	ID string
	// Name is the name of the registered domain.
	Name string
}

// Host defines mapping from the subdomain to the host IP addresses.
type Host struct {
	// ID is a unique identifier for a group of records. It is used as part
	// of the Pulumi resource ID and is expected to be human-readable and
	// descriptive enough in the configuration file.
	ID string
	// Name is the name of the host and subdomain.
	Name string
	// Addresses is a list of IP addresses of the host. A or AAAA record
	// will be created on the subdomain for each address in the list.
	Addresses []HostAddress
}

// HostAddress defines an A or AAAA DNS record.
type HostAddress struct {
	// ID is a unique identifier for the record. It is used as part of the
	// Pulumi resource ID and is expected to be human-readable and
	// descriptive enough in the configuration file.
	ID string
	// Value is the IP address of the host.
	Value net.IP
}

// Alias defines a CNAME DNS record.
type Alias struct {
	// ID is a unique identifier for the record. It is used as part of the
	// Pulumi resource ID and is expected to be human-readable and
	// descriptive enough in the configuration file.
	ID string
	// Name is the name of the DNS record. If empty, defaults to @ (root
	// domain).
	Name string
	// Host is the value for the CNAME DNS record.
	Host string
	// Proxied enables Cloudflare proxy for the DNS record.
	// See https://developers.cloudflare.com/dns/manage-dns-records/reference/proxied-dns-records
	Proxied bool
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

	zone, err := cloudflare.NewZone(ctx, resourceName,
		&cloudflare.ZoneArgs{
			Zone: pulumi.String(d.Name),
		},
		pulumi.DeleteBeforeReplace(true),
		pulumi.IgnoreChanges([]string{"accountId"}),
	)
	if err != nil {
		return err
	}

	if _, err := cloudflare.NewZoneSettingsOverride(ctx, resourceName,
		&cloudflare.ZoneSettingsOverrideArgs{
			ZoneId: zone.ID(),
			Settings: &cloudflare.ZoneSettingsOverrideSettingsArgs{
				Ssl:           pulumi.String("strict"),
				MinTlsVersion: pulumi.String("1.2"),
				UniversalSsl:  pulumi.String("on"),
			},
		},
		pulumi.DeleteBeforeReplace(true),
	); err != nil {
		return err
	}

	if err := setupHosts(ctx, zone, z, d); err != nil {
		return err
	}
	if err := setupAliases(ctx, zone, z, d); err != nil {
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
	resourceName := joinDash(z.ID, d.ID, host.ID, addr.ID)

	typ := "AAAA"
	if addr.Value.To4() != nil {
		typ = "A"
	}

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId: zone.ID(),
		Name:   pulumi.String(host.Name),
		Value:  pulumi.String(addr.Value.String()),
		Type:   pulumi.String(typ),
	})
	return err
}

func setupAliases(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain) error {
	for _, alias := range z.Aliases {
		if err := setupAlias(ctx, zone, z, d, &alias); err != nil {
			return err
		}
	}
	return nil
}

func setupAlias(ctx *pulumi.Context, zone *cloudflare.Zone, z *Zone, d *Domain, alias *Alias) error {
	resourceName := joinDash(z.ID, d.ID, alias.ID)

	name := alias.Name
	if name == "" {
		name = "@"
	}

	// Looks like the replacement we have to do here manually does not work
	// over API, although it is possible to set CNAME for subdomain in
	// Cloudflare dashboard.
	host := strings.ReplaceAll(alias.Host, "@", d.Name)

	_, err := cloudflare.NewRecord(ctx, resourceName, &cloudflare.RecordArgs{
		ZoneId:  zone.ID(),
		Name:    pulumi.String(name),
		Value:   pulumi.String(host),
		Proxied: pulumi.Bool(alias.Proxied),
		Type:    pulumi.String("CNAME"),
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
			Name:     pulumi.String(srv.Name),
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
