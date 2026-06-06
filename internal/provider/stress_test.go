//go:build stress

package provider

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func randSuffix(t *testing.T, n int) string {
	h := fnv.New64a()
	h.Write([]byte(t.Name()))
	src := rand.New(rand.NewPCG(h.Sum64(), 0))
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[src.IntN(len(chars))]
	}
	return string(b)
}

func requireWithin(t *testing.T, timeout time.Duration, name string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("%s did not complete within %s (server may be overloaded)", name, timeout)
	}
}

func stressBulkZonesConfig(prefix string, zoneCount, recordsPerZone int) string {
	var b strings.Builder
	b.WriteString("provider \"technitium\" {}\n\n")
	for i := 0; i < zoneCount; i++ {
		zoneName := fmt.Sprintf("%s-z%d.example", prefix, i)
		zoneRef := fmt.Sprintf("zone%d", i)
		fmt.Fprintf(&b, "resource \"technitium_dns_zone\" %q {\n  name = %q\n  type = \"Primary\"\n}\n\n",
			zoneRef, zoneName)
		for j := 0; j < recordsPerZone; j++ {
			recRef := fmt.Sprintf("rec%d_%d", i, j)
			domain := fmt.Sprintf("host%d.%s", j, zoneName)
			ip := fmt.Sprintf("10.%d.%d.%d", i, j/254, j%254+1)
			fmt.Fprintf(&b, "resource \"technitium_dns_record\" %q {\n"+
				"  zone   = technitium_dns_zone.%s.name\n"+
				"  domain = %q\n"+
				"  type   = \"A\"\n"+
				"  value  = %q\n"+
				"  ttl    = 300\n"+
				"}\n\n",
				recRef, zoneRef, domain, ip)
		}
	}
	return b.String()
}

func stressBulkDHCPScopesConfig(prefix string, count int) string {
	var b strings.Builder
	b.WriteString("provider \"technitium\" {}\n\n")
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("%s-dhcp%d", prefix, i)
		octet2 := 200 + i
		fmt.Fprintf(&b, "resource \"technitium_dhcp_scope\" %q {\n"+
			"  name          = %q\n"+
			"  start_address = \"10.%d.0.10\"\n"+
			"  end_address   = \"10.%d.0.200\"\n"+
			"  subnet_mask   = \"255.255.255.0\"\n"+
			"  router_address = \"10.%d.0.1\"\n"+
			"  enabled       = true\n"+
			"  lease_time    = 3600\n"+
			"}\n\n",
			fmt.Sprintf("scope%d", i), name, octet2, octet2, octet2)
	}
	return b.String()
}

func testAccCheckBulkZonesDestroy(prefix string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}
		for i := 0; i < count; i++ {
			zoneName := fmt.Sprintf("%s-z%d.example", prefix, i)
			_, err = c.GetZoneOptions(context.Background(), zoneName)
			if err == nil {
				return fmt.Errorf("zone %q still exists after destroy", zoneName)
			}
		}
		return nil
	}
}

func testAccCheckBulkDHCPScopesDestroy(prefix string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c, err := testAccClientFromEnv()
		if err != nil {
			return fmt.Errorf("creating client for destroy check: %s", err)
		}
		for i := 0; i < count; i++ {
			name := fmt.Sprintf("%s-dhcp%d", prefix, i)
			_, err = c.GetDHCPScope(context.Background(), name)
			if err == nil {
				return fmt.Errorf("DHCP scope %q still exists after destroy", name)
			}
		}
		return nil
	}
}

func TestAccStress_BulkZonesWithRecords(t *testing.T) {
	const zoneCount = 25
	const recordsPerZone = 4
	prefix := fmt.Sprintf("stress-%s", randSuffix(t, 6))

	requireWithin(t, 3*time.Minute, "bulk zones (25 zones, 100 records)", func() {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccCheckBulkZonesDestroy(prefix, zoneCount),
			Steps: []resource.TestStep{
				{
					Config: stressBulkZonesConfig(prefix, zoneCount, recordsPerZone),
				},
			},
		})
	})
}

func TestAccStress_BulkDHCPScopes(t *testing.T) {
	const count = 10
	prefix := fmt.Sprintf("stress-%s", randSuffix(t, 6))

	requireWithin(t, 2*time.Minute, "bulk DHCP scopes (10 scopes)", func() {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy:             testAccCheckBulkDHCPScopesDestroy(prefix, count),
			Steps: []resource.TestStep{
				{
					Config: stressBulkDHCPScopesConfig(prefix, count),
				},
			},
		})
	})
}

func TestAccStress_MixedResourceRefresh(t *testing.T) {
	const zoneCount = 20
	const recordsPerZone = 2
	const dhcpCount = 5
	prefix := fmt.Sprintf("stress-%s", randSuffix(t, 6))

	zonesConfig := stressBulkZonesConfig(prefix, zoneCount, recordsPerZone)
	dhcpConfig := stressBulkDHCPScopesConfig(prefix, dhcpCount)
	combined := zonesConfig + "\n" + strings.Replace(dhcpConfig, "provider \"technitium\" {}\n\n", "", 1)

	requireWithin(t, 4*time.Minute, "mixed resource refresh (65 resources)", func() {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
			CheckDestroy: func(s *terraform.State) error {
				if err := testAccCheckBulkZonesDestroy(prefix, zoneCount)(s); err != nil {
					return err
				}
				return testAccCheckBulkDHCPScopesDestroy(prefix, dhcpCount)(s)
			},
			Steps: []resource.TestStep{
				{
					Config: combined,
				},
				{
					Config:   combined,
					PlanOnly: true,
				},
			},
		})
	})
}
