package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &adminUserResource{}
	_ resource.ResourceWithImportState = &adminUserResource{}
)

type adminUserResource struct {
	client *client.Client
}

type adminUserResourceModel struct {
	Username              types.String `tfsdk:"username"`
	DisplayName           types.String `tfsdk:"display_name"`
	Password              types.String `tfsdk:"password"`
	Disabled              types.Bool   `tfsdk:"disabled"`
	SessionTimeoutSeconds types.Int64  `tfsdk:"session_timeout_seconds"`
	MemberOfGroups        types.String `tfsdk:"member_of_groups"`
}

func NewAdminUserResource() resource.Resource {
	return &adminUserResource{}
}

func (r *adminUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_user"
}

func (r *adminUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an admin user on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username for the admin user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password for the admin user. Write-only; not read back from the API.",
				Required:    true,
				Sensitive:   true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name for the admin user.",
				Optional:    true,
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the admin user is disabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"session_timeout_seconds": schema.Int64Attribute{
				Description: "Session timeout in seconds for the admin user.",
				Optional:    true,
				Computed:    true,
			},
			"member_of_groups": schema.StringAttribute{
				Description: "Comma-separated list of group names the user belongs to.",
				Optional:    true,
			},
		},
	}
}

func (r *adminUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adminUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan adminUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()
	password := plan.Password.ValueString()

	tflog.Debug(ctx, "Creating admin user", map[string]interface{}{
		"username": username,
	})

	extra := url.Values{}
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		extra.Set("displayName", plan.DisplayName.ValueString())
	}

	_, err := r.client.CreateUser(username, password, extra)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Admin User",
			fmt.Sprintf("Could not create user %q: %s", username, err),
		)
		return
	}

	// Set additional properties if specified.
	setParams := url.Values{}
	needsSet := false

	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() && plan.Disabled.ValueBool() {
		setParams.Set("disabled", "true")
		needsSet = true
	}

	if !plan.SessionTimeoutSeconds.IsNull() && !plan.SessionTimeoutSeconds.IsUnknown() {
		setParams.Set("sessionTimeoutSeconds", fmt.Sprintf("%d", plan.SessionTimeoutSeconds.ValueInt64()))
		needsSet = true
	}

	if !plan.MemberOfGroups.IsNull() && !plan.MemberOfGroups.IsUnknown() {
		setParams.Set("memberOfGroups", plan.MemberOfGroups.ValueString())
		needsSet = true
	}

	if needsSet {
		_, err := r.client.SetUserDetails(username, setParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Admin User Details",
				fmt.Sprintf("Could not set details for user %q: %s", username, err),
			)
			return
		}
	}

	// Read back.
	r.readUserIntoModel(username, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state adminUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()

	readResp, err := r.client.GetUserDetails(username)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateModelFromResponse(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *adminUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan adminUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()

	tflog.Debug(ctx, "Updating admin user", map[string]interface{}{
		"username": username,
	})

	params := url.Values{}

	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		params.Set("displayName", plan.DisplayName.ValueString())
	}

	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() {
		if plan.Disabled.ValueBool() {
			params.Set("disabled", "true")
		} else {
			params.Set("disabled", "false")
		}
	}

	if !plan.SessionTimeoutSeconds.IsNull() && !plan.SessionTimeoutSeconds.IsUnknown() {
		params.Set("sessionTimeoutSeconds", fmt.Sprintf("%d", plan.SessionTimeoutSeconds.ValueInt64()))
	}

	if !plan.MemberOfGroups.IsNull() && !plan.MemberOfGroups.IsUnknown() {
		params.Set("memberOfGroups", plan.MemberOfGroups.ValueString())
	}

	_, err := r.client.SetUserDetails(username, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Admin User",
			fmt.Sprintf("Could not update user %q: %s", username, err),
		)
		return
	}

	// Read back.
	r.readUserIntoModel(username, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state adminUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()

	tflog.Debug(ctx, "Deleting admin user", map[string]interface{}{
		"username": username,
	})

	_, err := r.client.DeleteUser(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Admin User",
			fmt.Sprintf("Could not delete user %q: %s", username, err),
		)
	}
}

func (r *adminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	username := req.ID

	readResp, err := r.client.GetUserDetails(username)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Admin User",
			fmt.Sprintf("Could not read user %q: %s", username, err),
		)
		return
	}

	model := adminUserResourceModel{
		Username: types.StringValue(username),
		Password: types.StringNull(),
	}

	r.populateModelFromResponse(readResp, &model)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readUserIntoModel reads user details from the API and populates the model.
// Password is preserved from the plan since it is write-only.
func (r *adminUserResource) readUserIntoModel(username string, model *adminUserResourceModel, diags *diag.Diagnostics) {
	readResp, err := r.client.GetUserDetails(username)
	if err != nil {
		diags.AddError(
			"Error Reading Admin User",
			fmt.Sprintf("Could not read user %q: %s", username, err),
		)
		return
	}

	// Preserve password since it's write-only.
	password := model.Password
	r.populateModelFromResponse(readResp, model)
	model.Password = password
}

// populateModelFromResponse fills the model fields from an API response map.
func (r *adminUserResource) populateModelFromResponse(resp map[string]interface{}, model *adminUserResourceModel) {
	if displayName, ok := resp["displayName"].(string); ok {
		model.DisplayName = types.StringValue(displayName)
	}

	if disabled, ok := resp["disabled"].(bool); ok {
		model.Disabled = types.BoolValue(disabled)
	}

	if timeout, ok := resp["sessionTimeoutSeconds"].(float64); ok {
		model.SessionTimeoutSeconds = types.Int64Value(int64(timeout))
	}

	if groups, ok := resp["memberOfGroups"].([]interface{}); ok {
		names := make([]string, 0, len(groups))
		for _, g := range groups {
			if gMap, ok := g.(map[string]interface{}); ok {
				if name, ok := gMap["name"].(string); ok {
					names = append(names, name)
				}
			} else if name, ok := g.(string); ok {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			model.MemberOfGroups = types.StringValue(strings.Join(names, ","))
		}
	}
}
