package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
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
	StartAddress     types.String `tfsdk:"start_address"`
	EndAddress       types.String `tfsdk:"end_address"`
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
						"start_address": schema.StringAttribute{
							Description: "The start IP address of the DHCP range.",
							Computed:    true,
						},
						"end_address": schema.StringAttribute{
							Description: "The end IP address of the DHCP range.",
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
	result, err := d.client.ListDHCPScopes(ctx)
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
		s, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := s["name"].(string)
		startAddr, _ := s["startingAddress"].(string)
		endAddr, _ := s["endingAddress"].(string)
		mask, _ := s["subnetMask"].(string)
		if name == "" {
			continue
		}
		scope := dhcpScopeDataModel{
			Name:         types.StringValue(name),
			StartAddress: types.StringValue(startAddr),
			EndAddress:   types.StringValue(endAddr),
			SubnetMask:   types.StringValue(mask),
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
