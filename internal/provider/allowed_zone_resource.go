package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &allowedZoneResource{}
	_ resource.ResourceWithImportState = &allowedZoneResource{}
)

type allowedZoneResource struct {
	client *client.Client
}

type allowedZoneResourceModel struct {
	Domain types.String `tfsdk:"domain"`
}

func NewAllowedZoneResource() resource.Resource {
	return &allowedZoneResource{}
}

func (r *allowedZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_allowed_zone"
}

func (r *allowedZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an entry in the Technitium DNS Server Allowed Zones list.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to add to the Allowed Zones.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *allowedZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *allowedZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan allowedZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	tflog.Debug(ctx, "Creating allowed zone", map[string]interface{}{
		"domain": domain,
	})

	_, err := r.client.AllowZone(domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Allowed Zone",
			fmt.Sprintf("Could not add %q to allowed zones: %s", domain, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *allowedZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state allowedZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	result, err := r.client.ListAllowedZones(domain)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	records, _ := result["records"].([]interface{})
	if len(records) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *allowedZoneResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update. Domain uses RequiresReplace, so any change triggers
	// delete + create.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Allowed zone resources do not support in-place updates. Changes to the domain trigger replacement.",
	)
}

func (r *allowedZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state allowedZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	tflog.Debug(ctx, "Deleting allowed zone", map[string]interface{}{
		"domain": domain,
	})

	_, err := r.client.DeleteAllowedZone(domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Allowed Zone",
			fmt.Sprintf("Could not remove %q from allowed zones: %s", domain, err),
		)
	}
}

func (r *allowedZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := allowedZoneResourceModel{
		Domain: types.StringValue(req.ID),
	}

	// Verify the zone exists.
	_, err := r.client.ListAllowedZones(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Allowed Zone",
			fmt.Sprintf("Could not find allowed zone %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
