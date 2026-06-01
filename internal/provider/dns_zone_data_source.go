package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &dnsZonesDataSource{}

func NewDNSZonesDataSource() datasource.DataSource {
	return &dnsZonesDataSource{}
}

type dnsZonesDataSource struct {
	client *client.Client
}

type dnsZonesDataSourceModel struct {
	Zones []dnsZoneModel `tfsdk:"zones"`
}

type dnsZoneModel struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Disabled     types.Bool   `tfsdk:"disabled"`
	DnssecStatus types.String `tfsdk:"dnssec_status"`
}

func (d *dnsZonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zones"
}

func (d *dnsZonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all authoritative DNS zones on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Description: "List of DNS zones.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The domain name of the zone.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of zone.",
							Computed:    true,
						},
						"disabled": schema.BoolAttribute{
							Description: "Whether the zone is disabled.",
							Computed:    true,
						},
						"dnssec_status": schema.StringAttribute{
							Description: "The DNSSEC status of the zone.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dnsZonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *dnsZonesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading DNS zones list")

	result, err := d.client.ListZones()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing DNS Zones",
			fmt.Sprintf("Could not list zones: %s", err),
		)
		return
	}

	var state dnsZonesDataSourceModel

	zoneList, _ := result["zones"].([]interface{})
	for _, entry := range zoneList {
		z, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		zone := dnsZoneModel{
			Name:         types.StringValue(stringFromMap(z, "name")),
			Type:         types.StringValue(stringFromMap(z, "type")),
			Disabled:     types.BoolValue(boolFromMap(z, "disabled")),
			DnssecStatus: types.StringValue(stringFromMap(z, "dnssecStatus")),
		}
		state.Zones = append(state.Zones, zone)
	}

	if state.Zones == nil {
		state.Zones = []dnsZoneModel{}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func stringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolFromMap(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
