package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium/internal/client"
)

var _ datasource.DataSource = &dhcpScopesDataSource{}

func NewDHCPScopesDataSource() datasource.DataSource {
	return &dhcpScopesDataSource{}
}

type dhcpScopesDataSource struct {
	client *client.Client
}

type dhcpScopesDataSourceModel struct {
	Scopes []dhcpScopeDataModel `tfsdk:"scopes"`
}

type dhcpScopeDataModel struct {
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	StartingAddress  types.String `tfsdk:"starting_address"`
	EndingAddress    types.String `tfsdk:"ending_address"`
	SubnetMask       types.String `tfsdk:"subnet_mask"`
	NetworkAddress   types.String `tfsdk:"network_address"`
	BroadcastAddress types.String `tfsdk:"broadcast_address"`
}

func (d *dhcpScopesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_scopes"
}

func (d *dhcpScopesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all DHCP scopes on the Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"scopes": schema.ListNestedAttribute{
				Description: "List of DHCP scopes.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the DHCP scope.",
							Computed:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Whether the DHCP scope is enabled.",
							Computed:    true,
						},
						"starting_address": schema.StringAttribute{
							Description: "The starting IP address of the DHCP range.",
							Computed:    true,
						},
						"ending_address": schema.StringAttribute{
							Description: "The ending IP address of the DHCP range.",
							Computed:    true,
						},
						"subnet_mask": schema.StringAttribute{
							Description: "The subnet mask for the scope.",
							Computed:    true,
						},
						"network_address": schema.StringAttribute{
							Description: "The network address of the scope.",
							Computed:    true,
						},
						"broadcast_address": schema.StringAttribute{
							Description: "The broadcast address of the scope.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dhcpScopesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dhcpScopesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.ListDHCPScopes()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DHCP Scopes",
			fmt.Sprintf("Could not list DHCP scopes: %s", err),
		)
		return
	}

	var state dhcpScopesDataSourceModel

	scopeList, ok := result["scopes"].([]interface{})
	if !ok {
		state.Scopes = []dhcpScopeDataModel{}
		diags := resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, entry := range scopeList {
		s := entry.(map[string]interface{})
		scope := dhcpScopeDataModel{
			Name:            types.StringValue(s["name"].(string)),
			StartingAddress: types.StringValue(s["startingAddress"].(string)),
			EndingAddress:   types.StringValue(s["endingAddress"].(string)),
			SubnetMask:      types.StringValue(s["subnetMask"].(string)),
		}

		if v, ok := s["enabled"].(bool); ok {
			scope.Enabled = types.BoolValue(v)
		} else {
			scope.Enabled = types.BoolValue(false)
		}

		if v, ok := s["networkAddress"].(string); ok {
			scope.NetworkAddress = types.StringValue(v)
		} else {
			scope.NetworkAddress = types.StringNull()
		}

		if v, ok := s["broadcastAddress"].(string); ok {
			scope.BroadcastAddress = types.StringValue(v)
		} else {
			scope.BroadcastAddress = types.StringNull()
		}

		state.Scopes = append(state.Scopes, scope)
	}

	if state.Scopes == nil {
		state.Scopes = []dhcpScopeDataModel{}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
