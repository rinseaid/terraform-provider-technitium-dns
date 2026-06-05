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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &dhcpReservedLeaseResource{}
	_ resource.ResourceWithImportState = &dhcpReservedLeaseResource{}
)

func NewDHCPReservedLeaseResource() resource.Resource {
	return &dhcpReservedLeaseResource{}
}

type dhcpReservedLeaseResource struct {
	client *client.Client
}

type dhcpReservedLeaseResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ScopeName       types.String `tfsdk:"scope_name"`
	HardwareAddress types.String `tfsdk:"hardware_address"`
	Address         types.String `tfsdk:"address"`
	Hostname        types.String `tfsdk:"hostname"`
	Comments        types.String `tfsdk:"comments"`
}

func (r *dhcpReservedLeaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcp_reserved_lease"
}

func (r *dhcpReservedLeaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a reserved DHCP lease in a Technitium DNS Server DHCP scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the format scope_name:hardware_address.",
				Computed:    true,
			},
			"scope_name": schema.StringAttribute{
				Description: "The name of the DHCP scope containing this reservation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hardware_address": schema.StringAttribute{
				Description: "The MAC address of the client device. Stored in lowercase colon-separated format.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					normalizeMACModifier{},
					stringplanmodifier.RequiresReplace(),
				},
			},
			"address": schema.StringAttribute{
				Description: "The reserved IP address for the client.",
				Required:    true,
			},
			"hostname": schema.StringAttribute{
				Description: "The hostname of the client.",
				Optional:    true,
			},
			"comments": schema.StringAttribute{
				Description: "Comments for this reserved lease entry.",
				Optional:    true,
			},
		},
	}
}

func (r *dhcpReservedLeaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dhcpReservedLeaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dhcpReservedLeaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeName := plan.ScopeName.ValueString()
	params := r.buildAddParams(&plan)

	tflog.Debug(ctx, "Creating DHCP reserved lease", map[string]interface{}{
		"scope":            scopeName,
		"hardware_address": plan.HardwareAddress.ValueString(),
	})

	_, err := r.client.AddReservedLease(scopeName, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DHCP Reserved Lease",
			fmt.Sprintf("Could not create reserved lease in scope %q: %s", scopeName, err),
		)
		return
	}

	plan.ID = types.StringValue(compositeLeaseID(scopeName, plan.HardwareAddress.ValueString()))

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dhcpReservedLeaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dhcpReservedLeaseResourceModel
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

func (r *dhcpReservedLeaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dhcpReservedLeaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeName := plan.ScopeName.ValueString()
	mac := plan.HardwareAddress.ValueString()

	tflog.Debug(ctx, "Updating DHCP reserved lease", map[string]interface{}{
		"scope":            scopeName,
		"hardware_address": mac,
	})

	// Remove the existing reservation and re-add with updated values.
	// The API does not support in-place updates for reserved leases.
	removeParams := url.Values{}
	removeParams.Set("hardwareAddress", mac)
	_, err := r.client.RemoveReservedLease(scopeName, removeParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DHCP Reserved Lease",
			fmt.Sprintf("Could not remove existing reservation for update in scope %q: %s", scopeName, err),
		)
		return
	}

	addParams := r.buildAddParams(&plan)
	_, err = r.client.AddReservedLease(scopeName, addParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DHCP Reserved Lease",
			fmt.Sprintf("Could not re-add reservation in scope %q: %s", scopeName, err),
		)
		return
	}

	plan.ID = types.StringValue(compositeLeaseID(scopeName, mac))

	diags = r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dhcpReservedLeaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dhcpReservedLeaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopeName := state.ScopeName.ValueString()
	mac := state.HardwareAddress.ValueString()

	tflog.Debug(ctx, "Deleting DHCP reserved lease", map[string]interface{}{
		"scope":            scopeName,
		"hardware_address": mac,
	})

	params := url.Values{}
	params.Set("hardwareAddress", mac)
	_, err := r.client.RemoveReservedLease(scopeName, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DHCP Reserved Lease",
			fmt.Sprintf("Could not remove reserved lease from scope %q: %s", scopeName, err),
		)
		return
	}
}

func (r *dhcpReservedLeaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in the format scope_name:hardware_address, got: %q", req.ID),
		)
		return
	}

	normalizedMAC := normalizeMAC(parts[1])
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), compositeLeaseID(parts[0], normalizedMAC))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hardware_address"), normalizedMAC)...)
}

func (r *dhcpReservedLeaseResource) buildAddParams(model *dhcpReservedLeaseResourceModel) url.Values {
	params := url.Values{}
	params.Set("hardwareAddress", model.HardwareAddress.ValueString())
	params.Set("ipAddress", model.Address.ValueString())

	if !model.Hostname.IsNull() && !model.Hostname.IsUnknown() {
		params.Set("hostName", model.Hostname.ValueString())
	}

	if !model.Comments.IsNull() && !model.Comments.IsUnknown() {
		params.Set("comments", model.Comments.ValueString())
	}

	return params
}

func (r *dhcpReservedLeaseResource) readIntoModel(_ context.Context, model *dhcpReservedLeaseResourceModel) (diags diag.Diagnostics) {
	scopeName := model.ScopeName.ValueString()
	mac := normalizeMAC(model.HardwareAddress.ValueString())

	scopeData, err := r.client.GetDHCPScope(scopeName)
	if err != nil {
		diags.AddError(
			"Error Reading DHCP Scope",
			fmt.Sprintf("Could not read scope %q to find reserved lease: %s", scopeName, err),
		)
		return
	}

	reservedLeases, ok := scopeData["reservedLeases"].([]interface{})
	if !ok {
		diags.AddError(
			"Error Reading DHCP Reserved Lease",
			fmt.Sprintf("No reserved leases found in scope %q", scopeName),
		)
		return
	}

	var found bool
	for _, entry := range reservedLeases {
		lease := entry.(map[string]interface{})
		leaseMAC := normalizeMAC(lease["hardwareAddress"].(string))
		if leaseMAC == mac {
			model.Address = types.StringValue(lease["address"].(string))

			if v, ok := lease["hostName"]; ok && v != nil && v != "" {
				model.Hostname = types.StringValue(v.(string))
			} else if !model.Hostname.IsNull() {
				model.Hostname = types.StringNull()
			}

			if v, ok := lease["comments"]; ok && v != nil && v != "" {
				model.Comments = types.StringValue(v.(string))
			} else if !model.Comments.IsNull() {
				model.Comments = types.StringNull()
			}

			found = true
			break
		}
	}

	if !found {
		diags.AddError(
			"Error Reading DHCP Reserved Lease",
			fmt.Sprintf("Reserved lease with MAC %q not found in scope %q", mac, scopeName),
		)
		return
	}

	model.ID = types.StringValue(compositeLeaseID(scopeName, model.HardwareAddress.ValueString()))

	return
}

func compositeLeaseID(scope, mac string) string {
	return scope + ":" + normalizeMAC(mac)
}

// normalizeMAC converts MAC addresses to a consistent lowercase colon-separated format.
func normalizeMAC(mac string) string {
	mac = strings.ToLower(mac)
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ReplaceAll(mac, ".", ":")
	return mac
}

type normalizeMACModifier struct{}

func (m normalizeMACModifier) Description(_ context.Context) string {
	return "Normalizes MAC address to lowercase colon-separated format."
}

func (m normalizeMACModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes MAC address to lowercase colon-separated format."
}

func (m normalizeMACModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	resp.PlanValue = types.StringValue(normalizeMAC(req.PlanValue.ValueString()))
}
