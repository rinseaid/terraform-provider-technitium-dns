package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium/internal/client"
)

var _ datasource.DataSource = &dhcpLeasesDataSource{}

func NewDHCPLeasesDataSource() datasource.DataSource {
	return &dhcpLeasesDataSource{}
}

type dhcpLeasesDataSource struct {
	client *client.Client
}

type dhcpLeasesDataSourceModel struct {
	ScopeName types.String       `tfsdk:"scope_name"`
	Leases    []dhcpLeaseModel   `tfsdk:"leases"`
}

type dhcpLeaseModel struct {
	Scope            types.String `tfsdk:"scope"`
	Type             types.String `tfsdk:"type"`
	HardwareAddress  types.String `tfsdk:"hardware_address"`
	ClientIdentifier types.String `tfsdk:"client_identifier"`
	Address          types.String `tfsdk:"address"`
	HostName         types.String `tfsdk:"hostname"`
	LeaseObtained    types.String `tfsdk:"lease_obtained"`
	LeaseExpires     types.String `tfsdk:"lease_expires"`
}

func (d *dhcpLeasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_leases"
}

func (d *dhcpLeasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all DHCP leases (dynamic and reserved) in a Technitium DNS Server DHCP scope.",
		Attributes: map[string]schema.Attribute{
			"scope_name": schema.StringAttribute{
				Description: "The name of the DHCP scope to list leases from.",
				Required:    true,
			},
			"leases": schema.ListNestedAttribute{
				Description: "List of DHCP leases in the scope.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"scope": schema.StringAttribute{
							Description: "The scope name for this lease.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The lease type (Dynamic or Reserved).",
							Computed:    true,
						},
						"hardware_address": schema.StringAttribute{
							Description: "The MAC address of the client.",
							Computed:    true,
						},
						"client_identifier": schema.StringAttribute{
							Description: "The client identifier.",
							Computed:    true,
						},
						"address": schema.StringAttribute{
							Description: "The leased IP address.",
							Computed:    true,
						},
						"hostname": schema.StringAttribute{
							Description: "The hostname of the client.",
							Computed:    true,
						},
						"lease_obtained": schema.StringAttribute{
							Description: "The timestamp when the lease was obtained.",
							Computed:    true,
						},
						"lease_expires": schema.StringAttribute{
							Description: "The timestamp when the lease expires.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dhcpLeasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dhcpLeasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dhcpLeasesDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeName := config.ScopeName.ValueString()

	result, err := d.client.ListDHCPLeases(scopeName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DHCP Leases",
			fmt.Sprintf("Could not list leases for scope %q: %s", scopeName, err),
		)
		return
	}

	var state dhcpLeasesDataSourceModel
	state.ScopeName = config.ScopeName

	leaseList, ok := result["leases"].([]interface{})
	if !ok {
		state.Leases = []dhcpLeaseModel{}
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		return
	}

	for _, entry := range leaseList {
		l := entry.(map[string]interface{})

		// Filter to only leases belonging to this scope
		leaseScope, _ := l["scope"].(string)
		if leaseScope != scopeName {
			continue
		}

		lease := dhcpLeaseModel{
			Scope: types.StringValue(leaseScope),
		}

		if v, ok := l["type"].(string); ok {
			lease.Type = types.StringValue(v)
		} else {
			lease.Type = types.StringNull()
		}

		if v, ok := l["hardwareAddress"].(string); ok {
			lease.HardwareAddress = types.StringValue(normalizeMAC(v))
		} else {
			lease.HardwareAddress = types.StringNull()
		}

		if v, ok := l["clientIdentifier"].(string); ok && v != "" {
			lease.ClientIdentifier = types.StringValue(v)
		} else {
			lease.ClientIdentifier = types.StringNull()
		}

		if v, ok := l["address"].(string); ok {
			lease.Address = types.StringValue(v)
		} else {
			lease.Address = types.StringNull()
		}

		if v, ok := l["hostName"].(string); ok && v != "" {
			lease.HostName = types.StringValue(v)
		} else {
			lease.HostName = types.StringNull()
		}

		if v, ok := l["leaseObtained"].(string); ok {
			lease.LeaseObtained = types.StringValue(v)
		} else {
			lease.LeaseObtained = types.StringNull()
		}

		if v, ok := l["leaseExpires"].(string); ok {
			lease.LeaseExpires = types.StringValue(v)
		} else {
			lease.LeaseExpires = types.StringNull()
		}

		state.Leases = append(state.Leases, lease)
	}

	if state.Leases == nil {
		state.Leases = []dhcpLeaseModel{}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
