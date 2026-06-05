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
	_ resource.Resource                = &dnsAppConfigResource{}
	_ resource.ResourceWithImportState = &dnsAppConfigResource{}
)

type dnsAppConfigResource struct {
	client *client.Client
}

type dnsAppConfigResourceModel struct {
	Name   types.String `tfsdk:"name"`
	Config types.String `tfsdk:"config"`
}

func NewDNSAppConfigResource() resource.Resource {
	return &dnsAppConfigResource{}
}

func (r *dnsAppConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_app_config"
}

func (r *dnsAppConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the configuration of an installed DNS app on a Technitium DNS Server. " +
			"The app must already be installed; this resource only manages its config string.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the DNS app.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				Description: "JSON configuration string for the DNS app.",
				Required:    true,
			},
		},
	}
}

func (r *dnsAppConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnsAppConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsAppConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	config := plan.Config.ValueString()

	tflog.Debug(ctx, "Setting DNS app config", map[string]interface{}{
		"name": name,
	})

	_, err := r.client.SetAppConfig(name, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting DNS App Config",
			fmt.Sprintf("Could not set config for app %q: %s", name, err),
		)
		return
	}

	// Read back the config.
	readResp, err := r.client.GetAppConfig(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DNS App Config",
			fmt.Sprintf("Could not read config for app %q after set: %s", name, err),
		)
		return
	}

	if cfg, ok := readResp["config"].(string); ok {
		plan.Config = types.StringValue(cfg)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsAppConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsAppConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	readResp, err := r.client.GetAppConfig(name)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if cfg, ok := readResp["config"].(string); ok {
		state.Config = types.StringValue(cfg)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsAppConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsAppConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	config := plan.Config.ValueString()

	tflog.Debug(ctx, "Updating DNS app config", map[string]interface{}{
		"name": name,
	})

	_, err := r.client.SetAppConfig(name, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DNS App Config",
			fmt.Sprintf("Could not set config for app %q: %s", name, err),
		)
		return
	}

	// Read back.
	readResp, err := r.client.GetAppConfig(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DNS App Config",
			fmt.Sprintf("Could not read config for app %q after update: %s", name, err),
		)
		return
	}

	if cfg, ok := readResp["config"].(string); ok {
		plan.Config = types.StringValue(cfg)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsAppConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: removing the config from state only. The app remains installed
	// with its existing config.
	tflog.Debug(ctx, "Removing DNS app config from state (no-op on server)")
}

func (r *dnsAppConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	readResp, err := r.client.GetAppConfig(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing DNS App Config",
			fmt.Sprintf("Could not read config for app %q: %s", name, err),
		)
		return
	}

	model := dnsAppConfigResourceModel{
		Name: types.StringValue(name),
	}

	if cfg, ok := readResp["config"].(string); ok {
		model.Config = types.StringValue(cfg)
	} else {
		model.Config = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
