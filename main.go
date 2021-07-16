package main

import (
	"strings"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(run)
}

func run(ctx *pulumi.Context) error {
	if err := setupAddressZone(ctx); err != nil {
		return err
	}
	if err := setupAliasedZone(ctx); err != nil {
		return err
	}
	return nil
}

func setupAliasedZone(ctx *pulumi.Context) error {
	zone, err := cloudflareZone(ctx, "tie.wtf")
	if err != nil {
		return err
	}

	records := []struct {
		ID    string
		Name  string
		Value string
	}{
		{"873d10f6", "roku", "roku.tie.rip."},
		{"d85ffc21", "ubernet", "ubernet.tie.rip."},
		{"ab9173dd", "madara", "madara.tie.rip."},
		{"8b913246", "saitama", "saitama.tie.rip."},
		{"eaed6781", "tatsuya", "tatsuya.tie.rip."},
		{"eaed6781", "brim", "brim.ml."},
	}

	for _, r := range records {
		_, err := cloudflare.NewRecord(ctx, r.ID, &cloudflare.RecordArgs{
			ZoneId: zone.ID(),
			Name:   pulumi.String(r.Name),
			Value:  pulumi.String(r.Value),
			Type:   pulumi.String("CNAME"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func setupAddressZone(ctx *pulumi.Context) error {
	zone, err := cloudflareZone(ctx, "tie.rip")
	if err != nil {
		return err
	}

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
		{"93cba5bd", "tatsuya", "37.110.66.21"},
		{"f7bf04d2", "tatsuya", prefixISP + ":2"},
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

func cloudflareZone(ctx *pulumi.Context, domain string) (*cloudflare.Zone, error) {
	return cloudflare.NewZone(ctx, domain, &cloudflare.ZoneArgs{
		Zone: pulumi.String(domain),
	})
}
