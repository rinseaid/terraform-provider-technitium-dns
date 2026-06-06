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
	_ resource.Resource                = &catalogZoneMembershipResource{}
	_ resource.ResourceWithImportState = &catalogZoneMembershipResource{}
)

func NewCatalogZoneMembershipResource() resource.Resource {
	return &catalogZoneMembershipResource{}
}

type catalogZoneMembershipResource struct {
	client *client.Client
}

type catalogZoneMembershipResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	CatalogZone types.String `tfsdk:"catalog_zone"`
}

func (r *catalogZoneMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog_zone_membership"
}

func (r *catalogZoneMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages catalog zone membership for a DNS zone on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for the resource. Same as the zone name.",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "The domain name of the member zone.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"catalog_zone": schema.StringAttribute{
				Description: "The catalog zone to assign this zone to.",
				Required:    true,
			},
		},
	}
}

func (r *catalogZoneMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *catalogZoneMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan catalogZoneMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := plan.Zone.ValueString()
	catalogZone := plan.CatalogZone.ValueString()

	tflog.Debug(ctx, "Setting catalog zone membership", map[string]interface{}{
		"zone":         zoneName,
		"catalog_zone": catalogZone,
	})

	params := url.Values{}
	params.Set("catalog", catalogZone)
	_, err := r.client.SetZoneOptions(ctx, zoneName, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Catalog Zone Membership",
			fmt.Sprintf("Could not assign zone %q to catalog %q: %s", zoneName, catalogZone, err),
		)
		return
	}

	diags := r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *catalogZoneMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state catalogZoneMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readIntoModel(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *catalogZoneMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan catalogZoneMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := plan.Zone.ValueString()
	catalogZone := plan.CatalogZone.ValueString()

	tflog.Debug(ctx, "Updating catalog zone membership", map[string]interface{}{
		"zone":         zoneName,
		"catalog_zone": catalogZone,
	})

	params := url.Values{}
	params.Set("catalog", catalogZone)
	_, err := r.client.SetZoneOptions(ctx, zoneName, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Catalog Zone Membership",
			fmt.Sprintf("Could not update zone %q catalog to %q: %s", zoneName, catalogZone, err),
		)
		return
	}

	diags := r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *catalogZoneMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state catalogZoneMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := state.Zone.ValueString()

	tflog.Debug(ctx, "Removing catalog zone membership", map[string]interface{}{
		"zone": zoneName,
	})

	params := url.Values{}
	params.Set("catalog", "")
	_, err := r.client.SetZoneOptions(ctx, zoneName, params)
	if err != nil {
		// If the zone no longer exists (deleted out-of-band), treat as success.
		if strings.Contains(err.Error(), "was not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Removing Catalog Zone Membership",
			fmt.Sprintf("Could not remove catalog membership from zone %q: %s", zoneName, err),
		)
	}
}

func (r *catalogZoneMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("zone"), req, resp)
}

// readIntoModel fetches zone options and populates the catalog membership fields.
func (r *catalogZoneMembershipResource) readIntoModel(ctx context.Context, model *catalogZoneMembershipResourceModel) (diags diag.Diagnostics) {
	zoneName := model.Zone.ValueString()

	tflog.Debug(ctx, "Reading catalog zone membership", map[string]interface{}{"zone": zoneName})

	response, err := r.client.GetZoneOptions(ctx, zoneName)
	if err != nil {
		diags.AddError(
			"Error Reading Catalog Zone Membership",
			fmt.Sprintf("Could not read zone options for %q: %s", zoneName, err),
		)
		return
	}

	catalog, ok := response["catalog"].(string)
	if !ok || catalog == "" {
		diags.AddError(
			"Catalog Zone Membership Not Found",
			fmt.Sprintf("Zone %q is not a member of any catalog zone.", zoneName),
		)
		return
	}

	model.ID = types.StringValue(zoneName)
	model.Zone = types.StringValue(zoneName)
	model.CatalogZone = types.StringValue(catalog)

	return
}
