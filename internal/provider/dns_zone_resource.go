package provider

import (
	"context"
	"fmt"

	"net/url"

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
	"github.com/rinseaid/terraform-provider-technitium/internal/client"
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
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Disabled     types.Bool   `tfsdk:"disabled"`
	DnssecStatus types.String `tfsdk:"dnssec_status"`
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
				Description: "The type of zone. Valid values: Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Primary",
						"Secondary",
						"Stub",
						"Forwarder",
						"SecondaryForwarder",
						"Catalog",
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

	_, err := r.client.CreateZone(plan.Name.ValueString(), plan.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Zone",
			fmt.Sprintf("Could not create zone %q: %s", plan.Name.ValueString(), err),
		)
		return
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

	// Update zone type if changed.
	if !plan.Type.Equal(state.Type) {
		params := url.Values{}
		params.Set("type", plan.Type.ValueString())
		_, err := r.client.SetZoneOptions(zoneName, params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating DNS Zone Type",
				fmt.Sprintf("Could not update zone %q type: %s", zoneName, err),
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

	return
}
