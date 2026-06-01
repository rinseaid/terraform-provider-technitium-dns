package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium/internal/client"
)

var (
	_ resource.Resource                = &dhcpScopeResource{}
	_ resource.ResourceWithImportState = &dhcpScopeResource{}
)

func NewDHCPScopeResource() resource.Resource {
	return &dhcpScopeResource{}
}

type dhcpScopeResource struct {
	client *client.Client
}

type dhcpScopeResourceModel struct {
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	StartAddress     types.String `tfsdk:"start_address"`
	EndAddress       types.String `tfsdk:"end_address"`
	SubnetMask       types.String `tfsdk:"subnet_mask"`
	RouterAddress    types.String `tfsdk:"router_address"`
	DNSServers       types.List   `tfsdk:"dns_servers"`
	DomainName       types.String `tfsdk:"domain_name"`
	LeaseTime        types.Int64  `tfsdk:"lease_time"`
	OfferDelay       types.Int64  `tfsdk:"offer_delay"`
	PingCheck        types.Bool   `tfsdk:"ping_check"`
	PingCheckTimeout types.Int64  `tfsdk:"ping_check_timeout"`
}

func (r *dhcpScopeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_scope"
}

func (r *dhcpScopeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Technitium DNS Server DHCP scope.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the DHCP scope.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the DHCP scope is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"start_address": schema.StringAttribute{
				Description: "The starting IP address of the DHCP range.",
				Required:    true,
			},
			"end_address": schema.StringAttribute{
				Description: "The ending IP address of the DHCP range.",
				Required:    true,
			},
			"subnet_mask": schema.StringAttribute{
				Description: "The subnet mask for the DHCP scope.",
				Required:    true,
			},
			"router_address": schema.StringAttribute{
				Description: "The default gateway IP address for clients.",
				Optional:    true,
			},
			"dns_servers": schema.ListAttribute{
				Description: "List of DNS server IP addresses for clients.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name for this network.",
				Optional:    true,
			},
			"lease_time": schema.Int64Attribute{
				Description: "Lease duration in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			"offer_delay": schema.Int64Attribute{
				Description: "Delay before sending DHCPOFFER in milliseconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"ping_check": schema.BoolAttribute{
				Description: "Enable ping check before assigning an IP address.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ping_check_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for ping check.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
			},
		},
	}
}

func (r *dhcpScopeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *dhcpScopeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildSetParams(ctx, &plan)

	tflog.Debug(ctx, "Creating DHCP scope", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.SetDHCPScope(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DHCP Scope",
			fmt.Sprintf("Could not create scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	if plan.Enabled.ValueBool() {
		_, err = r.client.EnableDHCPScope(plan.Name.ValueString())
	} else {
		_, err = r.client.DisableDHCPScope(plan.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting DHCP Scope Enabled State",
			fmt.Sprintf("Could not set enabled state for scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dhcpScopeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpScopeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.readIntoModel(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *dhcpScopeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildSetParams(ctx, &plan)

	tflog.Debug(ctx, "Updating DHCP scope", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.SetDHCPScope(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DHCP Scope",
			fmt.Sprintf("Could not update scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	if plan.Enabled.ValueBool() {
		_, err = r.client.EnableDHCPScope(plan.Name.ValueString())
	} else {
		_, err = r.client.DisableDHCPScope(plan.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting DHCP Scope Enabled State",
			fmt.Sprintf("Could not set enabled state for scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dhcpScopeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpScopeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting DHCP scope", map[string]interface{}{"name": state.Name.ValueString()})

	_, err := r.client.DeleteDHCPScope(state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DHCP Scope",
			fmt.Sprintf("Could not delete scope %q: %s", state.Name.ValueString(), err),
		)
		return
	}
}

func (r *dhcpScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *dhcpScopeResource) buildSetParams(ctx context.Context, model *dhcpScopeResourceModel) url.Values {
	params := url.Values{}
	params.Set("name", model.Name.ValueString())
	params.Set("startingAddress", model.StartAddress.ValueString())
	params.Set("endingAddress", model.EndAddress.ValueString())
	params.Set("subnetMask", model.SubnetMask.ValueString())

	if !model.RouterAddress.IsNull() && !model.RouterAddress.IsUnknown() {
		params.Set("routerAddress", model.RouterAddress.ValueString())
	}

	if !model.DNSServers.IsNull() && !model.DNSServers.IsUnknown() {
		var servers []string
		model.DNSServers.ElementsAs(ctx, &servers, false)
		params.Set("dnsServers", strings.Join(servers, ","))
	}

	if !model.DomainName.IsNull() && !model.DomainName.IsUnknown() {
		params.Set("domainName", model.DomainName.ValueString())
	}

	// Convert lease_time (seconds) to days/hours/minutes for API
	leaseSeconds := model.LeaseTime.ValueInt64()
	days := leaseSeconds / 86400
	remainder := leaseSeconds % 86400
	hours := remainder / 3600
	remainder = remainder % 3600
	minutes := remainder / 60
	params.Set("leaseTimeDays", fmt.Sprintf("%d", days))
	params.Set("leaseTimeHours", fmt.Sprintf("%d", hours))
	params.Set("leaseTimeMinutes", fmt.Sprintf("%d", minutes))

	params.Set("offerDelayTime", fmt.Sprintf("%d", model.OfferDelay.ValueInt64()))
	params.Set("pingCheckEnabled", fmt.Sprintf("%t", model.PingCheck.ValueBool()))

	if !model.PingCheckTimeout.IsNull() && !model.PingCheckTimeout.IsUnknown() {
		params.Set("pingCheckTimeout", fmt.Sprintf("%d", model.PingCheckTimeout.ValueInt64()))
	}

	return params
}

func (r *dhcpScopeResource) readIntoModel(ctx context.Context, model *dhcpScopeResourceModel) (diags diag.Diagnostics) {
	scopeData, err := r.client.GetDHCPScope(model.Name.ValueString())
	if err != nil {
		diags.AddError(
			"Error Reading DHCP Scope",
			fmt.Sprintf("Could not read scope %q: %s", model.Name.ValueString(), err),
		)
		return
	}

	model.Name = types.StringValue(scopeData["name"].(string))
	model.StartAddress = types.StringValue(scopeData["startingAddress"].(string))
	model.EndAddress = types.StringValue(scopeData["endingAddress"].(string))
	model.SubnetMask = types.StringValue(scopeData["subnetMask"].(string))

	if v, ok := scopeData["routerAddress"]; ok && v != nil && v != "" {
		model.RouterAddress = types.StringValue(v.(string))
	} else if !model.RouterAddress.IsNull() {
		model.RouterAddress = types.StringNull()
	}

	if v, ok := scopeData["dnsServers"]; ok && v != nil {
		servers := v.([]interface{})
		serverStrings := make([]string, len(servers))
		for i, s := range servers {
			serverStrings[i] = s.(string)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, serverStrings)
		diags.Append(d...)
		model.DNSServers = listVal
	} else if !model.DNSServers.IsNull() {
		model.DNSServers = types.ListNull(types.StringType)
	}

	if v, ok := scopeData["domainName"]; ok && v != nil && v != "" {
		model.DomainName = types.StringValue(v.(string))
	} else if !model.DomainName.IsNull() {
		model.DomainName = types.StringNull()
	}

	// Convert days/hours/minutes back to total seconds
	var totalSeconds int64
	if v, ok := scopeData["leaseTimeDays"]; ok && v != nil {
		totalSeconds += int64(v.(float64)) * 86400
	}
	if v, ok := scopeData["leaseTimeHours"]; ok && v != nil {
		totalSeconds += int64(v.(float64)) * 3600
	}
	if v, ok := scopeData["leaseTimeMinutes"]; ok && v != nil {
		totalSeconds += int64(v.(float64)) * 60
	}
	model.LeaseTime = types.Int64Value(totalSeconds)

	if v, ok := scopeData["offerDelayTime"]; ok && v != nil {
		model.OfferDelay = types.Int64Value(int64(v.(float64)))
	}

	if v, ok := scopeData["pingCheckEnabled"]; ok && v != nil {
		model.PingCheck = types.BoolValue(v.(bool))
	}

	if v, ok := scopeData["pingCheckTimeout"]; ok && v != nil {
		model.PingCheckTimeout = types.Int64Value(int64(v.(float64)))
	} else if !model.PingCheckTimeout.IsNull() {
		model.PingCheckTimeout = types.Int64Null()
	}

	// Determine enabled state from list endpoint (GetDHCPScope does not include it)
	scopes, err := r.client.ListDHCPScopes()
	if err != nil {
		diags.AddError(
			"Error Reading DHCP Scopes List",
			fmt.Sprintf("Could not list scopes to determine enabled state: %s", err),
		)
		return
	}

	model.Enabled = types.BoolValue(false)
	if scopeList, ok := scopes["scopes"].([]interface{}); ok {
		for _, s := range scopeList {
			scopeMap := s.(map[string]interface{})
			if scopeMap["name"].(string) == model.Name.ValueString() {
				if enabled, ok := scopeMap["enabled"].(bool); ok {
					model.Enabled = types.BoolValue(enabled)
				}
				break
			}
		}
	}

	return
}
