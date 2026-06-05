package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &dnsZoneResource{}
	_ resource.ResourceWithImportState = &dnsZoneResource{}
)

func NewDNSZoneResource() resource.Resource {
	return &dnsZoneResource{}
}

type dnsZoneResource struct {
	client *client.Client
}

type dnsZoneResourceModel struct {
	Name                        types.String `tfsdk:"name"`
	Type                        types.String `tfsdk:"type"`
	Disabled                    types.Bool   `tfsdk:"disabled"`
	DnssecStatus                types.String `tfsdk:"dnssec_status"`
	ZoneTransfer                types.String `tfsdk:"zone_transfer"`
	Notify                      types.String `tfsdk:"notify"`
	NotifyNameServers           types.String `tfsdk:"notify_name_servers"`
	ZoneTransferNetworkACL      types.String `tfsdk:"zone_transfer_network_acl"`
	ZoneTransferTSIGKeyNames    types.String `tfsdk:"zone_transfer_tsig_key_names"`
	QueryAccess                 types.String `tfsdk:"query_access"`
	Update                      types.String `tfsdk:"update"`
	PrimaryNameServerAddresses  types.String `tfsdk:"primary_name_server_addresses"`
	PrimaryZoneTransferProtocol types.String `tfsdk:"primary_zone_transfer_protocol"`
	PrimaryZoneTransferTSIGKey  types.String `tfsdk:"primary_zone_transfer_tsig_key"`
}

func (r *dnsZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (r *dnsZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an authoritative DNS zone on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The domain name of the zone.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of zone. Valid values: Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog, SecondaryCatalog.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Primary",
						"Secondary",
						"Stub",
						"Forwarder",
						"SecondaryForwarder",
						"Catalog",
						"SecondaryCatalog",
					),
				},
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the zone is disabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dnssec_status": schema.StringAttribute{
				Description: "The DNSSEC status of the zone.",
				Computed:    true,
			},
			"zone_transfer": schema.StringAttribute{
				Description: "Zone transfer policy. Valid values: Deny, Allow, AllowOnlyZoneNameServers, UseSpecifiedNetworkACL, AllowZoneNameServersAndUseSpecifiedNetworkACL.",
				Optional:    true,
				Computed:    true,
			},
			"notify": schema.StringAttribute{
				Description: "Zone update notification policy. Valid values: None, ZoneNameServers, SpecifiedNameServers, BothZoneAndSpecifiedNameServers.",
				Optional:    true,
				Computed:    true,
			},
			"notify_name_servers": schema.StringAttribute{
				Description: "Comma-separated list of IP addresses to notify for zone updates.",
				Optional:    true,
			},
			"zone_transfer_network_acl": schema.StringAttribute{
				Description: "Comma-separated network ACL for zone transfers. Prefix with ! to deny.",
				Optional:    true,
			},
			"zone_transfer_tsig_key_names": schema.StringAttribute{
				Description: "Comma-separated TSIG key names authorized for zone transfers.",
				Optional:    true,
			},
			"query_access": schema.StringAttribute{
				Description: "Query access policy. Valid values: Deny, Allow, AllowOnlyPrivateNetworks, AllowOnlyZoneNameServers, UseSpecifiedNetworkACL, AllowZoneNameServersAndUseSpecifiedNetworkACL.",
				Optional:    true,
				Computed:    true,
			},
			"update": schema.StringAttribute{
				Description: "Dynamic update policy (RFC 2136). Valid values: Deny, Allow, AllowOnlyZoneNameServers, UseSpecifiedNetworkACL, AllowZoneNameServersAndUseSpecifiedNetworkACL.",
				Optional:    true,
				Computed:    true,
			},
			"primary_name_server_addresses": schema.StringAttribute{
				Description: "Comma-separated primary name server addresses for Secondary, SecondaryForwarder, SecondaryCatalog, and Stub zones.",
				Optional:    true,
			},
			"primary_zone_transfer_protocol": schema.StringAttribute{
				Description: "Zone transfer protocol for Secondary, SecondaryForwarder, and SecondaryCatalog zones. Valid values: Tcp, Tls, Quic.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Tcp", "Tls", "Quic"),
				},
			},
			"primary_zone_transfer_tsig_key": schema.StringAttribute{
				Description: "TSIG key name for zone transfers on Secondary, SecondaryForwarder, and SecondaryCatalog zones.",
				Optional:    true,
			},
		},
	}
}

func (r *dnsZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnsZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating DNS zone", map[string]interface{}{
		"zone": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	})

	var createExtra url.Values
	zoneType := plan.Type.ValueString()
	if zoneType == "Forwarder" || zoneType == "SecondaryForwarder" {
		createExtra = url.Values{}
		createExtra.Set("forwarder", "this-server")
		createExtra.Set("protocol", "Udp")
	}
	if zoneType == "Secondary" || zoneType == "SecondaryForwarder" || zoneType == "SecondaryCatalog" || zoneType == "Stub" {
		if createExtra == nil {
			createExtra = url.Values{}
		}
		if !plan.PrimaryNameServerAddresses.IsNull() && !plan.PrimaryNameServerAddresses.IsUnknown() {
			createExtra.Set("primaryNameServerAddresses", plan.PrimaryNameServerAddresses.ValueString())
		}
		if !plan.PrimaryZoneTransferProtocol.IsNull() && !plan.PrimaryZoneTransferProtocol.IsUnknown() {
			createExtra.Set("zoneTransferProtocol", plan.PrimaryZoneTransferProtocol.ValueString())
		}
		if !plan.PrimaryZoneTransferTSIGKey.IsNull() && !plan.PrimaryZoneTransferTSIGKey.IsUnknown() {
			createExtra.Set("tsigKeyName", plan.PrimaryZoneTransferTSIGKey.ValueString())
		}
	}
	_, err := r.client.CreateZone(plan.Name.ValueString(), zoneType, createExtra)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Zone",
			fmt.Sprintf("Could not create zone %q: %s", plan.Name.ValueString(), err),
		)
		return
	}

	// Set zone options if any were specified.
	opts := url.Values{}
	hasOpts := false
	if !plan.ZoneTransfer.IsNull() && !plan.ZoneTransfer.IsUnknown() {
		opts.Set("zoneTransfer", plan.ZoneTransfer.ValueString())
		hasOpts = true
	}
	if !plan.Notify.IsNull() && !plan.Notify.IsUnknown() {
		opts.Set("notify", plan.Notify.ValueString())
		hasOpts = true
	}
	if !plan.NotifyNameServers.IsNull() && !plan.NotifyNameServers.IsUnknown() {
		opts.Set("notifyNameServers", plan.NotifyNameServers.ValueString())
		hasOpts = true
	}
	if !plan.ZoneTransferNetworkACL.IsNull() && !plan.ZoneTransferNetworkACL.IsUnknown() {
		opts.Set("zoneTransferNetworkACL", plan.ZoneTransferNetworkACL.ValueString())
		hasOpts = true
	}
	if !plan.ZoneTransferTSIGKeyNames.IsNull() && !plan.ZoneTransferTSIGKeyNames.IsUnknown() {
		opts.Set("zoneTransferTsigKeyNames", plan.ZoneTransferTSIGKeyNames.ValueString())
		hasOpts = true
	}
	if !plan.QueryAccess.IsNull() && !plan.QueryAccess.IsUnknown() {
		opts.Set("queryAccess", plan.QueryAccess.ValueString())
		hasOpts = true
	}
	if !plan.Update.IsNull() && !plan.Update.IsUnknown() {
		opts.Set("update", plan.Update.ValueString())
		hasOpts = true
	}
	if hasOpts {
		_, err = r.client.SetZoneOptions(plan.Name.ValueString(), opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting DNS Zone Options",
				fmt.Sprintf("Zone %q was created but options could not be set: %s", plan.Name.ValueString(), err),
			)
			return
		}
	}

	// Zones are created enabled by default. Disable if requested.
	if plan.Disabled.ValueBool() {
		_, err = r.client.DisableZone(plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Disabling DNS Zone",
				fmt.Sprintf("Zone %q was created but could not be disabled: %s", plan.Name.ValueString(), err),
			)
			return
		}
	}

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.readIntoModel(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsZoneResourceModel
	var state dnsZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := plan.Name.ValueString()

	tflog.Debug(ctx, "Updating DNS zone", map[string]interface{}{"zone": zoneName})

	// Update zone options if any changed.
	params := url.Values{}
	needsUpdate := false

	if !plan.Type.Equal(state.Type) {
		params.Set("type", plan.Type.ValueString())
		needsUpdate = true
	}
	if !plan.ZoneTransfer.Equal(state.ZoneTransfer) && !plan.ZoneTransfer.IsNull() {
		params.Set("zoneTransfer", plan.ZoneTransfer.ValueString())
		needsUpdate = true
	}
	if !plan.Notify.Equal(state.Notify) && !plan.Notify.IsNull() {
		params.Set("notify", plan.Notify.ValueString())
		needsUpdate = true
	}
	if !plan.NotifyNameServers.Equal(state.NotifyNameServers) {
		if !plan.NotifyNameServers.IsNull() {
			params.Set("notifyNameServers", plan.NotifyNameServers.ValueString())
		} else {
			params.Set("notifyNameServers", "false")
		}
		needsUpdate = true
	}
	if !plan.ZoneTransferNetworkACL.Equal(state.ZoneTransferNetworkACL) {
		if !plan.ZoneTransferNetworkACL.IsNull() {
			params.Set("zoneTransferNetworkACL", plan.ZoneTransferNetworkACL.ValueString())
		} else {
			params.Set("zoneTransferNetworkACL", "false")
		}
		needsUpdate = true
	}
	if !plan.ZoneTransferTSIGKeyNames.Equal(state.ZoneTransferTSIGKeyNames) {
		if !plan.ZoneTransferTSIGKeyNames.IsNull() {
			params.Set("zoneTransferTsigKeyNames", plan.ZoneTransferTSIGKeyNames.ValueString())
		} else {
			params.Set("zoneTransferTsigKeyNames", "false")
		}
		needsUpdate = true
	}
	if !plan.QueryAccess.Equal(state.QueryAccess) && !plan.QueryAccess.IsNull() {
		params.Set("queryAccess", plan.QueryAccess.ValueString())
		needsUpdate = true
	}
	if !plan.Update.Equal(state.Update) && !plan.Update.IsNull() {
		params.Set("update", plan.Update.ValueString())
		needsUpdate = true
	}
	if !plan.PrimaryNameServerAddresses.Equal(state.PrimaryNameServerAddresses) {
		if !plan.PrimaryNameServerAddresses.IsNull() {
			params.Set("primaryNameServerAddresses", plan.PrimaryNameServerAddresses.ValueString())
		} else {
			params.Set("primaryNameServerAddresses", "false")
		}
		needsUpdate = true
	}
	if !plan.PrimaryZoneTransferProtocol.Equal(state.PrimaryZoneTransferProtocol) {
		if !plan.PrimaryZoneTransferProtocol.IsNull() {
			params.Set("primaryZoneTransferProtocol", plan.PrimaryZoneTransferProtocol.ValueString())
		}
		needsUpdate = true
	}
	if !plan.PrimaryZoneTransferTSIGKey.Equal(state.PrimaryZoneTransferTSIGKey) {
		if !plan.PrimaryZoneTransferTSIGKey.IsNull() {
			params.Set("primaryZoneTransferTsigKeyName", plan.PrimaryZoneTransferTSIGKey.ValueString())
		} else {
			params.Set("primaryZoneTransferTsigKeyName", "false")
		}
		needsUpdate = true
	}

	if needsUpdate {
		_, err := r.client.SetZoneOptions(zoneName, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating DNS Zone Options",
				fmt.Sprintf("Could not update zone %q options: %s", zoneName, err),
			)
			return
		}
	}

	// Update disabled state if changed.
	if !plan.Disabled.Equal(state.Disabled) {
		if plan.Disabled.ValueBool() {
			_, err := r.client.DisableZone(zoneName)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Disabling DNS Zone",
					fmt.Sprintf("Could not disable zone %q: %s", zoneName, err),
				)
				return
			}
		} else {
			_, err := r.client.EnableZone(zoneName)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Enabling DNS Zone",
					fmt.Sprintf("Could not enable zone %q: %s", zoneName, err),
				)
				return
			}
		}
	}

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting DNS zone", map[string]interface{}{"zone": state.Name.ValueString()})

	_, err := r.client.DeleteZone(state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "was not found") || strings.Contains(err.Error(), "No such zone") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting DNS Zone",
			fmt.Sprintf("Could not delete zone %q: %s", state.Name.ValueString(), err),
		)
	}
}

func (r *dnsZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// readIntoModel fetches zone options from the API and populates the model fields.
func (r *dnsZoneResource) readIntoModel(ctx context.Context, model *dnsZoneResourceModel) (diags diag.Diagnostics) {
	zoneName := model.Name.ValueString()

	tflog.Debug(ctx, "Reading DNS zone", map[string]interface{}{"zone": zoneName})

	response, err := r.client.GetZoneOptions(zoneName)
	if err != nil {
		diags.AddError(
			"Error Reading DNS Zone",
			fmt.Sprintf("Could not read zone %q: %s", zoneName, err),
		)
		return
	}

	if response == nil {
		diags.AddError(
			"Error Reading DNS Zone",
			fmt.Sprintf("Unexpected response format when reading zone %q.", zoneName),
		)
		return
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if zoneType, ok := response["type"].(string); ok {
		model.Type = types.StringValue(zoneType)
	}
	if disabled, ok := response["disabled"].(bool); ok {
		model.Disabled = types.BoolValue(disabled)
	}
	if dnssecStatus, ok := response["dnssecStatus"].(string); ok {
		model.DnssecStatus = types.StringValue(dnssecStatus)
	} else {
		model.DnssecStatus = types.StringValue("Unsigned")
	}

	if v, ok := response["zoneTransfer"].(string); ok {
		model.ZoneTransfer = types.StringValue(v)
	} else {
		model.ZoneTransfer = types.StringValue("Deny")
	}
	if v, ok := response["notify"].(string); ok {
		model.Notify = types.StringValue(v)
	} else {
		model.Notify = types.StringValue("None")
	}
	if v := joinStringList(response["notifyNameServers"]); v != "" {
		model.NotifyNameServers = types.StringValue(v)
	} else if !model.NotifyNameServers.IsNull() {
		model.NotifyNameServers = types.StringNull()
	}
	if v := joinStringList(response["zoneTransferNetworkACL"]); v != "" {
		model.ZoneTransferNetworkACL = types.StringValue(v)
	} else if !model.ZoneTransferNetworkACL.IsNull() {
		model.ZoneTransferNetworkACL = types.StringNull()
	}
	if v := joinStringList(response["zoneTransferTsigKeyNames"]); v != "" {
		model.ZoneTransferTSIGKeyNames = types.StringValue(v)
	} else if !model.ZoneTransferTSIGKeyNames.IsNull() {
		model.ZoneTransferTSIGKeyNames = types.StringNull()
	}
	if v, ok := response["queryAccess"].(string); ok {
		model.QueryAccess = types.StringValue(v)
	} else {
		model.QueryAccess = types.StringValue("Allow")
	}
	if v, ok := response["update"].(string); ok {
		model.Update = types.StringValue(v)
	} else {
		model.Update = types.StringValue("Deny")
	}

	if v := joinStringList(response["primaryNameServerAddresses"]); v != "" {
		model.PrimaryNameServerAddresses = types.StringValue(v)
	} else if !model.PrimaryNameServerAddresses.IsNull() {
		model.PrimaryNameServerAddresses = types.StringNull()
	}
	if v, ok := response["primaryZoneTransferProtocol"].(string); ok && v != "" {
		model.PrimaryZoneTransferProtocol = types.StringValue(v)
	} else if !model.PrimaryZoneTransferProtocol.IsNull() {
		model.PrimaryZoneTransferProtocol = types.StringNull()
	}
	if v, ok := response["primaryZoneTransferTsigKeyName"].(string); ok && v != "" {
		model.PrimaryZoneTransferTSIGKey = types.StringValue(v)
	} else if !model.PrimaryZoneTransferTSIGKey.IsNull() {
		model.PrimaryZoneTransferTSIGKey = types.StringNull()
	}

	return
}

func joinStringList(v interface{}) string {
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return ""
	}
	strs := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok && s != "" {
			strs = append(strs, s)
		}
	}
	return strings.Join(strs, ",")
}
