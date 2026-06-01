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
	_ resource.Resource                = &blockedZoneResource{}
	_ resource.ResourceWithImportState = &blockedZoneResource{}
)

type blockedZoneResource struct {
	client *client.Client
}

type blockedZoneResourceModel struct {
	Domain types.String `tfsdk:"domain"`
}

func NewBlockedZoneResource() resource.Resource {
	return &blockedZoneResource{}
}

func (r *blockedZoneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blocked_zone"
}

func (r *blockedZoneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an entry in the Technitium DNS Server Blocked Zones list.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The domain name to add to the Blocked Zones.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *blockedZoneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *blockedZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blockedZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	tflog.Debug(ctx, "Creating blocked zone", map[string]interface{}{
		"domain": domain,
	})

	_, err := r.client.BlockZone(domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Blocked Zone",
			fmt.Sprintf("Could not add %q to blocked zones: %s", domain, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *blockedZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blockedZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	// List the zone to verify it exists. The API returns records for the
	// domain when queried directly. An error indicates the zone is absent.
	_, err := r.client.ListBlockedZones(domain)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *blockedZoneResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update. Domain uses RequiresReplace, so any change triggers
	// delete + create.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Blocked zone resources do not support in-place updates. Changes to the domain trigger replacement.",
	)
}

func (r *blockedZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blockedZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()

	tflog.Debug(ctx, "Deleting blocked zone", map[string]interface{}{
		"domain": domain,
	})

	_, err := r.client.DeleteBlockedZone(domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Blocked Zone",
			fmt.Sprintf("Could not remove %q from blocked zones: %s", domain, err),
		)
	}
}

func (r *blockedZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := blockedZoneResourceModel{
		Domain: types.StringValue(req.ID),
	}

	// Verify the zone exists.
	_, err := r.client.ListBlockedZones(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Blocked Zone",
			fmt.Sprintf("Could not find blocked zone %q: %s", req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
