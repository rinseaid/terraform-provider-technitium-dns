package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &adminPermissionResource{}
	_ resource.ResourceWithImportState = &adminPermissionResource{}
)

type adminPermissionResource struct {
	client *client.Client
}

type adminPermissionResourceModel struct {
	Section          types.String `tfsdk:"section"`
	UserPermissions  types.String `tfsdk:"user_permissions"`
	GroupPermissions types.String `tfsdk:"group_permissions"`
}

func NewAdminPermissionResource() resource.Resource {
	return &adminPermissionResource{}
}

func (r *adminPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_permission"
}

func (r *adminPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages admin permissions for a section on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"section": schema.StringAttribute{
				Description: "The permission section name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Dashboard",
						"Zones",
						"Cache",
						"Allowed",
						"Blocked",
						"Apps",
						"DnsClient",
						"Settings",
						"DhcpServer",
						"Administration",
						"Logs",
					),
				},
			},
			"user_permissions": schema.StringAttribute{
				Description: "Pipe-delimited user permissions: username|canView|canModify|canDelete. " +
					"Multiple entries separated by pipes.",
				Optional: true,
			},
			"group_permissions": schema.StringAttribute{
				Description: "Pipe-delimited group permissions: name|canView|canModify|canDelete. " +
					"Multiple entries separated by pipes.",
				Optional: true,
			},
		},
	}
}

func (r *adminPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adminPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan adminPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	section := plan.Section.ValueString()

	tflog.Debug(ctx, "Setting admin permissions", map[string]interface{}{
		"section": section,
	})

	params := url.Values{}
	if !plan.UserPermissions.IsNull() && !plan.UserPermissions.IsUnknown() {
		params.Set("userPermissions", plan.UserPermissions.ValueString())
	}
	if !plan.GroupPermissions.IsNull() && !plan.GroupPermissions.IsUnknown() {
		params.Set("groupPermissions", plan.GroupPermissions.ValueString())
	}

	_, err := r.client.SetPermissions(section, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting Admin Permissions",
			fmt.Sprintf("Could not set permissions for section %q: %s", section, err),
		)
		return
	}

	// Read back.
	readResp, err := r.client.GetPermissions(section)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Admin Permissions",
			fmt.Sprintf("Could not read permissions for section %q: %s", section, err),
		)
		return
	}

	r.populateModelFromResponse(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state adminPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	section := state.Section.ValueString()

	readResp, err := r.client.GetPermissions(section)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateModelFromResponse(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *adminPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan adminPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	section := plan.Section.ValueString()

	tflog.Debug(ctx, "Updating admin permissions", map[string]interface{}{
		"section": section,
	})

	params := url.Values{}
	if !plan.UserPermissions.IsNull() && !plan.UserPermissions.IsUnknown() {
		params.Set("userPermissions", plan.UserPermissions.ValueString())
	}
	if !plan.GroupPermissions.IsNull() && !plan.GroupPermissions.IsUnknown() {
		params.Set("groupPermissions", plan.GroupPermissions.ValueString())
	}

	_, err := r.client.SetPermissions(section, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Admin Permissions",
			fmt.Sprintf("Could not update permissions for section %q: %s", section, err),
		)
		return
	}

	// Read back.
	readResp, err := r.client.GetPermissions(section)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Admin Permissions",
			fmt.Sprintf("Could not read permissions for section %q after update: %s", section, err),
		)
		return
	}

	r.populateModelFromResponse(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No-op: permissions cannot be deleted from the server, just removed from
	// Terraform state.
	tflog.Debug(ctx, "Removing admin permissions from state (no-op on server)")
}

func (r *adminPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	section := req.ID

	readResp, err := r.client.GetPermissions(section)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Admin Permissions",
			fmt.Sprintf("Could not read permissions for section %q: %s", section, err),
		)
		return
	}

	model := adminPermissionResourceModel{
		Section: types.StringValue(section),
	}

	r.populateModelFromResponse(readResp, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// populateModelFromResponse fills the model from an API response, reconstructing
// the pipe-delimited permission strings.
func (r *adminPermissionResource) populateModelFromResponse(resp map[string]interface{}, model *adminPermissionResourceModel) {
	if userPerms, ok := resp["userPermissions"].([]interface{}); ok && len(userPerms) > 0 {
		parts := make([]string, 0, len(userPerms)*4)
		for _, up := range userPerms {
			pMap, ok := up.(map[string]interface{})
			if !ok {
				continue
			}
			username, _ := pMap["username"].(string)
			canView := boolToString(pMap["canView"])
			canModify := boolToString(pMap["canModify"])
			canDelete := boolToString(pMap["canDelete"])
			parts = append(parts, username, canView, canModify, canDelete)
		}
		if len(parts) > 0 {
			model.UserPermissions = types.StringValue(strings.Join(parts, "|"))
		}
	}

	if groupPerms, ok := resp["groupPermissions"].([]interface{}); ok && len(groupPerms) > 0 {
		parts := make([]string, 0, len(groupPerms)*4)
		for _, gp := range groupPerms {
			pMap, ok := gp.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := pMap["name"].(string)
			canView := boolToString(pMap["canView"])
			canModify := boolToString(pMap["canModify"])
			canDelete := boolToString(pMap["canDelete"])
			parts = append(parts, name, canView, canModify, canDelete)
		}
		if len(parts) > 0 {
			model.GroupPermissions = types.StringValue(strings.Join(parts, "|"))
		}
	}
}

func boolToString(v interface{}) string {
	if b, ok := v.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return "false"
}
