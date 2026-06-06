package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &allowedZonesDataSource{}

type allowedZonesDataSource struct {
	client *client.Client
}

type allowedZonesDataSourceModel struct {
	Zones []allowedZoneModel `tfsdk:"zones"`
}

type allowedZoneModel struct {
	Domain types.String `tfsdk:"domain"`
}

func NewAllowedZonesDataSource() datasource.DataSource {
	return &allowedZonesDataSource{}
}

func (d *allowedZonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowed_zones"
}

func (d *allowedZonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all allowed zones on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Description: "List of allowed zones.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The domain name of the allowed zone.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *allowedZonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *allowedZonesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.ListAllowedZones(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Allowed Zones",
			fmt.Sprintf("Could not list allowed zones: %s", err),
		)
		return
	}

	var state allowedZonesDataSourceModel

	// The API returns a "zones" array containing objects with a "name" field
	// for each top-level zone entry.
	zoneList, ok := result["zones"].([]interface{})
	if ok {
		for _, entry := range zoneList {
			if z, ok := entry.(map[string]interface{}); ok {
				if name, ok := z["name"].(string); ok {
					state.Zones = append(state.Zones, allowedZoneModel{
						Domain: types.StringValue(name),
					})
				}
			}
		}
	}

	if state.Zones == nil {
		state.Zones = []allowedZoneModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
