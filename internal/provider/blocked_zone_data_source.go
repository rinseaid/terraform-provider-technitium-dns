package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &blockedZonesDataSource{}

type blockedZonesDataSource struct {
	client *client.Client
}

type blockedZonesDataSourceModel struct {
	Zones []blockedZoneModel `tfsdk:"zones"`
}

type blockedZoneModel struct {
	Domain types.String `tfsdk:"domain"`
}

func NewBlockedZonesDataSource() datasource.DataSource {
	return &blockedZonesDataSource{}
}

func (d *blockedZonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blocked_zones"
}

func (d *blockedZonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all blocked zones on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Description: "List of blocked zones.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The domain name of the blocked zone.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *blockedZonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *blockedZonesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.ListBlockedZones(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Blocked Zones",
			fmt.Sprintf("Could not list blocked zones: %s", err),
		)
		return
	}

	var state blockedZonesDataSourceModel

	// The API returns a "zones" array containing objects with a "name" field
	// for each top-level zone entry.
	zoneList, ok := result["zones"].([]interface{})
	if ok {
		for _, entry := range zoneList {
			if z, ok := entry.(map[string]interface{}); ok {
				if name, ok := z["name"].(string); ok {
					state.Zones = append(state.Zones, blockedZoneModel{
						Domain: types.StringValue(name),
					})
				}
			}
		}
	}

	if state.Zones == nil {
		state.Zones = []blockedZoneModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
