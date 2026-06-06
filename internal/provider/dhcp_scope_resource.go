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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
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
	Name                         types.String `tfsdk:"name"`
	Enabled                      types.Bool   `tfsdk:"enabled"`
	StartAddress                 types.String `tfsdk:"start_address"`
	EndAddress                   types.String `tfsdk:"end_address"`
	SubnetMask                   types.String `tfsdk:"subnet_mask"`
	RouterAddress                types.String `tfsdk:"router_address"`
	DNSServers                   types.List   `tfsdk:"dns_servers"`
	DomainName                   types.String `tfsdk:"domain_name"`
	LeaseTime                    types.Int64  `tfsdk:"lease_time"`
	OfferDelay                   types.Int64  `tfsdk:"offer_delay"`
	PingCheck                    types.Bool   `tfsdk:"ping_check"`
	PingCheckTimeout             types.Int64  `tfsdk:"ping_check_timeout"`
	PingCheckRetries             types.Int64  `tfsdk:"ping_check_retries"`
	DomainSearchList             types.List   `tfsdk:"domain_search_list"`
	DNSUpdates                   types.Bool   `tfsdk:"dns_updates"`
	DNSTTL                       types.Int64  `tfsdk:"dns_ttl"`
	BootFileName                 types.String `tfsdk:"boot_file_name"`
	TFTPServerAddresses          types.List   `tfsdk:"tftp_server_addresses"`
	Exclusions                   types.String `tfsdk:"exclusions"`
	NTPServers                   types.List   `tfsdk:"ntp_servers"`
	StaticRoutes                 types.String `tfsdk:"static_routes"`
	VendorInfo                   types.String `tfsdk:"vendor_info"`
	IgnoreClientIdentifierOption types.Bool   `tfsdk:"ignore_client_identifier_option"`
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
				Description: "Lease duration in seconds. Must be divisible by 60 (API resolution is minutes).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
				Validators: []validator.Int64{
					int64DivisibleBy(60),
				},
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
			"ping_check_retries": schema.Int64Attribute{
				Description: "Maximum number of ping requests to try before assigning an IP address.",
				Optional:    true,
				Computed:    true,
			},
			"domain_search_list": schema.ListAttribute{
				Description: "List of domain names clients can use as search suffixes (DHCP option 119).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"dns_updates": schema.BoolAttribute{
				Description: "Allow the DHCP server to automatically update forward and reverse DNS entries for clients.",
				Optional:    true,
				Computed:    true,
			},
			"dns_ttl": schema.Int64Attribute{
				Description: "TTL value in seconds for auto-created forward and reverse DNS records.",
				Optional:    true,
				Computed:    true,
			},
			"boot_file_name": schema.StringAttribute{
				Description: "Boot file name on the TFTP server for PXE clients (DHCP option 67).",
				Optional:    true,
			},
			"tftp_server_addresses": schema.ListAttribute{
				Description: "List of TFTP server addresses for PXE boot (DHCP option 150).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"exclusions": schema.StringAttribute{
				Description: "IP address ranges to exclude from dynamic assignment. Pipe-delimited pairs: startIP|endIP|startIP|endIP.",
				Optional:    true,
			},
			"ntp_servers": schema.ListAttribute{
				Description: "List of NTP server IP addresses for clients (DHCP option 42).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"static_routes": schema.StringAttribute{
				Description: "Pipe-delimited static routes: destination|subnetMask|gateway|destination|subnetMask|gateway (DHCP option 121).",
				Optional:    true,
			},
			"vendor_info": schema.StringAttribute{
				Description: "Pipe-delimited vendor information: classIdentifier|specificInfo|classIdentifier|specificInfo.",
				Optional:    true,
			},
			"ignore_client_identifier_option": schema.BoolAttribute{
				Description: "Always use client MAC address as identifier instead of Client Identifier (option 61).",
				Optional:    true,
				Computed:    true,
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

	params, paramDiags := r.buildSetParams(ctx, &plan)
	resp.Diagnostics.Append(paramDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DHCP scope", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.SetDHCPScope(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DHCP Scope",
			fmt.Sprintf("Could not create scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	if plan.Enabled.ValueBool() {
		_, err = r.client.EnableDHCPScope(ctx, plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Enabling DHCP Scope",
				fmt.Sprintf("Could not enable scope %q: %s", plan.Name.ValueString(), err),
			)
			return
		}
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
	var state dhcpScopeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params, paramDiags := r.buildSetParams(ctx, &plan)
	resp.Diagnostics.Append(paramDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating DHCP scope", map[string]interface{}{"name": plan.Name.ValueString()})

	_, err := r.client.SetDHCPScope(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DHCP Scope",
			fmt.Sprintf("Could not update scope %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	if !plan.Enabled.Equal(state.Enabled) {
		if plan.Enabled.ValueBool() {
			_, err = r.client.EnableDHCPScope(ctx, plan.Name.ValueString())
		} else {
			_, err = r.client.DisableDHCPScope(ctx, plan.Name.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting DHCP Scope Enabled State",
				fmt.Sprintf("Could not set enabled state for scope %q: %s", plan.Name.ValueString(), err),
			)
			return
		}
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

	_, err := r.client.DeleteDHCPScope(ctx, state.Name.ValueString())
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

func (r *dhcpScopeResource) buildSetParams(ctx context.Context, model *dhcpScopeResourceModel) (url.Values, diag.Diagnostics) {
	var diags diag.Diagnostics
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
		diags.Append(model.DNSServers.ElementsAs(ctx, &servers, false)...)
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
	remainder %= 3600
	minutes := remainder / 60
	params.Set("leaseTimeDays", fmt.Sprintf("%d", days))
	params.Set("leaseTimeHours", fmt.Sprintf("%d", hours))
	params.Set("leaseTimeMinutes", fmt.Sprintf("%d", minutes))

	params.Set("offerDelayTime", fmt.Sprintf("%d", model.OfferDelay.ValueInt64()))
	params.Set("pingCheckEnabled", fmt.Sprintf("%t", model.PingCheck.ValueBool()))

	if !model.PingCheckTimeout.IsNull() && !model.PingCheckTimeout.IsUnknown() {
		params.Set("pingCheckTimeout", fmt.Sprintf("%d", model.PingCheckTimeout.ValueInt64()))
	}

	if !model.PingCheckRetries.IsNull() && !model.PingCheckRetries.IsUnknown() {
		params.Set("pingCheckRetries", fmt.Sprintf("%d", model.PingCheckRetries.ValueInt64()))
	}

	if !model.DomainSearchList.IsNull() && !model.DomainSearchList.IsUnknown() {
		var domains []string
		diags.Append(model.DomainSearchList.ElementsAs(ctx, &domains, false)...)
		params.Set("domainSearchList", strings.Join(domains, ","))
	}

	if !model.DNSUpdates.IsNull() && !model.DNSUpdates.IsUnknown() {
		params.Set("dnsUpdates", fmt.Sprintf("%t", model.DNSUpdates.ValueBool()))
	}

	if !model.DNSTTL.IsNull() && !model.DNSTTL.IsUnknown() {
		params.Set("dnsTtl", fmt.Sprintf("%d", model.DNSTTL.ValueInt64()))
	}

	if !model.BootFileName.IsNull() && !model.BootFileName.IsUnknown() {
		params.Set("bootFileName", model.BootFileName.ValueString())
	}

	if !model.TFTPServerAddresses.IsNull() && !model.TFTPServerAddresses.IsUnknown() {
		var addrs []string
		diags.Append(model.TFTPServerAddresses.ElementsAs(ctx, &addrs, false)...)
		params.Set("tftpServerAddresses", strings.Join(addrs, ","))
	}

	if !model.Exclusions.IsNull() && !model.Exclusions.IsUnknown() {
		params.Set("exclusions", model.Exclusions.ValueString())
	}

	if !model.NTPServers.IsNull() && !model.NTPServers.IsUnknown() {
		var servers []string
		diags.Append(model.NTPServers.ElementsAs(ctx, &servers, false)...)
		params.Set("ntpServers", strings.Join(servers, ","))
	}

	if !model.StaticRoutes.IsNull() && !model.StaticRoutes.IsUnknown() {
		params.Set("staticRoutes", model.StaticRoutes.ValueString())
	}

	if !model.VendorInfo.IsNull() && !model.VendorInfo.IsUnknown() {
		params.Set("vendorInfo", model.VendorInfo.ValueString())
	}

	if !model.IgnoreClientIdentifierOption.IsNull() && !model.IgnoreClientIdentifierOption.IsUnknown() {
		params.Set("ignoreClientIdentifierOption", fmt.Sprintf("%t", model.IgnoreClientIdentifierOption.ValueBool()))
	}

	return params, diags
}

func (r *dhcpScopeResource) readIntoModel(ctx context.Context, model *dhcpScopeResourceModel) (diags diag.Diagnostics) {
	scopeData, err := r.client.GetDHCPScope(ctx, model.Name.ValueString())
	if err != nil {
		diags.AddError(
			"Error Reading DHCP Scope",
			fmt.Sprintf("Could not read scope %q: %s", model.Name.ValueString(), err),
		)
		return
	}

	if v, ok := scopeData["name"].(string); ok {
		model.Name = types.StringValue(v)
	}
	if v, ok := scopeData["startingAddress"].(string); ok {
		model.StartAddress = types.StringValue(v)
	}
	if v, ok := scopeData["endingAddress"].(string); ok {
		model.EndAddress = types.StringValue(v)
	}
	if v, ok := scopeData["subnetMask"].(string); ok {
		model.SubnetMask = types.StringValue(v)
	}

	if v, ok := scopeData["routerAddress"].(string); ok && v != "" {
		model.RouterAddress = types.StringValue(v)
	} else if !model.RouterAddress.IsNull() {
		model.RouterAddress = types.StringNull()
	}

	if v, ok := scopeData["dnsServers"]; ok && v != nil {
		servers, ok := v.([]interface{})
		if !ok {
			servers = nil
		}
		serverStrings := make([]string, 0, len(servers))
		for _, s := range servers {
			if str, ok := s.(string); ok {
				serverStrings = append(serverStrings, str)
			}
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, serverStrings)
		diags.Append(d...)
		model.DNSServers = listVal
	} else if !model.DNSServers.IsNull() {
		model.DNSServers = types.ListNull(types.StringType)
	}

	if v, ok := scopeData["domainName"].(string); ok && v != "" {
		model.DomainName = types.StringValue(v)
	} else if !model.DomainName.IsNull() {
		model.DomainName = types.StringNull()
	}

	// Convert days/hours/minutes back to total seconds
	var totalSeconds int64
	if v, ok := scopeData["leaseTimeDays"].(float64); ok {
		totalSeconds += int64(v) * 86400
	}
	if v, ok := scopeData["leaseTimeHours"].(float64); ok {
		totalSeconds += int64(v) * 3600
	}
	if v, ok := scopeData["leaseTimeMinutes"].(float64); ok {
		totalSeconds += int64(v) * 60
	}
	model.LeaseTime = types.Int64Value(totalSeconds)

	if v, ok := scopeData["offerDelayTime"].(float64); ok {
		model.OfferDelay = types.Int64Value(int64(v))
	}

	if v, ok := scopeData["pingCheckEnabled"].(bool); ok {
		model.PingCheck = types.BoolValue(v)
	}

	if v, ok := scopeData["pingCheckTimeout"].(float64); ok {
		model.PingCheckTimeout = types.Int64Value(int64(v))
	} else if !model.PingCheckTimeout.IsNull() {
		model.PingCheckTimeout = types.Int64Null()
	}

	if v, ok := scopeData["pingCheckRetries"].(float64); ok {
		model.PingCheckRetries = types.Int64Value(int64(v))
	} else if !model.PingCheckRetries.IsNull() {
		model.PingCheckRetries = types.Int64Null()
	}

	if v, ok := scopeData["domainSearchList"]; ok && v != nil {
		items, _ := v.([]interface{})
		strs := make([]string, 0, len(items))
		for _, s := range items {
			if str, ok := s.(string); ok {
				strs = append(strs, str)
			}
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, strs)
		diags.Append(d...)
		model.DomainSearchList = listVal
	} else if !model.DomainSearchList.IsNull() {
		model.DomainSearchList = types.ListNull(types.StringType)
	}

	if v, ok := scopeData["dnsUpdates"].(bool); ok {
		model.DNSUpdates = types.BoolValue(v)
	} else if !model.DNSUpdates.IsNull() {
		model.DNSUpdates = types.BoolNull()
	}

	if v, ok := scopeData["dnsTtl"].(float64); ok {
		model.DNSTTL = types.Int64Value(int64(v))
	} else if !model.DNSTTL.IsNull() {
		model.DNSTTL = types.Int64Null()
	}

	if v, ok := scopeData["bootFileName"].(string); ok && v != "" {
		model.BootFileName = types.StringValue(v)
	} else if !model.BootFileName.IsNull() {
		model.BootFileName = types.StringNull()
	}

	if v, ok := scopeData["tftpServerAddresses"]; ok && v != nil {
		items, _ := v.([]interface{})
		strs := make([]string, 0, len(items))
		for _, s := range items {
			if str, ok := s.(string); ok {
				strs = append(strs, str)
			}
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, strs)
		diags.Append(d...)
		model.TFTPServerAddresses = listVal
	} else if !model.TFTPServerAddresses.IsNull() {
		model.TFTPServerAddresses = types.ListNull(types.StringType)
	}

	if v, ok := scopeData["exclusions"]; ok && v != nil {
		if ranges, ok := v.([]interface{}); ok && len(ranges) > 0 {
			var parts []string
			for _, r := range ranges {
				rMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				start, _ := rMap["startingAddress"].(string)
				end, _ := rMap["endingAddress"].(string)
				if start != "" && end != "" {
					parts = append(parts, start, end)
				}
			}
			if len(parts) > 0 {
				model.Exclusions = types.StringValue(strings.Join(parts, "|"))
			} else {
				model.Exclusions = types.StringNull()
			}
		} else if !model.Exclusions.IsNull() {
			model.Exclusions = types.StringNull()
		}
	} else if !model.Exclusions.IsNull() {
		model.Exclusions = types.StringNull()
	}

	if v, ok := scopeData["ntpServers"]; ok && v != nil {
		items, _ := v.([]interface{})
		strs := make([]string, 0, len(items))
		for _, s := range items {
			if str, ok := s.(string); ok {
				strs = append(strs, str)
			}
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, strs)
		diags.Append(d...)
		model.NTPServers = listVal
	} else if !model.NTPServers.IsNull() {
		model.NTPServers = types.ListNull(types.StringType)
	}

	if v, ok := scopeData["staticRoutes"]; ok && v != nil {
		if routes, ok := v.([]interface{}); ok && len(routes) > 0 {
			var parts []string
			for _, r := range routes {
				rMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				dest, _ := rMap["destination"].(string)
				mask, _ := rMap["subnetMask"].(string)
				router, _ := rMap["router"].(string)
				if dest != "" && mask != "" && router != "" {
					parts = append(parts, dest, mask, router)
				}
			}
			if len(parts) > 0 {
				model.StaticRoutes = types.StringValue(strings.Join(parts, "|"))
			} else {
				model.StaticRoutes = types.StringNull()
			}
		} else if !model.StaticRoutes.IsNull() {
			model.StaticRoutes = types.StringNull()
		}
	} else if !model.StaticRoutes.IsNull() {
		model.StaticRoutes = types.StringNull()
	}

	if v, ok := scopeData["vendorInfo"]; ok && v != nil {
		if vendors, ok := v.([]interface{}); ok && len(vendors) > 0 {
			var parts []string
			for _, vi := range vendors {
				vMap, ok := vi.(map[string]interface{})
				if !ok {
					continue
				}
				classID, _ := vMap["identifier"].(string)
				info, _ := vMap["information"].(string)
				if classID != "" && info != "" {
					parts = append(parts, classID, info)
				}
			}
			if len(parts) > 0 {
				model.VendorInfo = types.StringValue(strings.Join(parts, "|"))
			} else {
				model.VendorInfo = types.StringNull()
			}
		} else if !model.VendorInfo.IsNull() {
			model.VendorInfo = types.StringNull()
		}
	} else if !model.VendorInfo.IsNull() {
		model.VendorInfo = types.StringNull()
	}

	if v, ok := scopeData["ignoreClientIdentifierOption"].(bool); ok {
		model.IgnoreClientIdentifierOption = types.BoolValue(v)
	} else if !model.IgnoreClientIdentifierOption.IsNull() {
		model.IgnoreClientIdentifierOption = types.BoolNull()
	}

	// Determine enabled state from list endpoint (GetDHCPScope does not include it)
	scopes, err := r.client.ListDHCPScopes(ctx)
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
			scopeMap, ok := s.(map[string]interface{})
			if !ok {
				continue
			}
			scopeName, ok := scopeMap["name"].(string)
			if !ok {
				continue
			}
			if scopeName == model.Name.ValueString() {
				if enabled, ok := scopeMap["enabled"].(bool); ok {
					model.Enabled = types.BoolValue(enabled)
				}
				break
			}
		}
	}

	return
}

type int64DivisibleByValidator struct {
	divisor int64
}

func int64DivisibleBy(divisor int64) int64DivisibleByValidator {
	return int64DivisibleByValidator{divisor: divisor}
}

func (v int64DivisibleByValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be divisible by %d", v.divisor)
}

func (v int64DivisibleByValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v int64DivisibleByValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueInt64()
	if val%v.divisor != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			fmt.Sprintf("Value %d must be divisible by %d.", val, v.divisor),
		)
	}
}
