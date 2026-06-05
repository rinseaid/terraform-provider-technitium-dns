package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &adminGroupResource{}
	_ resource.ResourceWithImportState = &adminGroupResource{}
)

type adminGroupResource struct {
	client *client.Client
}

type adminGroupResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Members     types.String `tfsdk:"members"`
}

func NewAdminGroupResource() resource.Resource {
	return &adminGroupResource{}
}

func (r *adminGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_group"
}

func (r *adminGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an admin group on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the admin group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the admin group.",
				Optional:    true,
			},
			"members": schema.StringAttribute{
				Description: "Comma-separated list of usernames that are members of this group.",
				Optional:    true,
			},
		},
	}
}

func (r *adminGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adminGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan adminGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	description := ""
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		description = plan.Description.ValueString()
	}

	tflog.Debug(ctx, "Creating admin group", map[string]interface{}{
		"name": name,
	})

	_, err := r.client.CreateGroup(name, description)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Admin Group",
			fmt.Sprintf("Could not create group %q: %s", name, err),
		)
		return
	}

	// Set members if specified.
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() && plan.Members.ValueString() != "" {
		setParams := url.Values{}
		setParams.Set("members", plan.Members.ValueString())
		_, err := r.client.SetGroupDetails(name, setParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Admin Group Members",
				fmt.Sprintf("Could not set members for group %q: %s", name, err),
			)
			return
		}
	}

	// Read back.
	readResp, err := r.client.GetGroupDetails(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Admin Group",
			fmt.Sprintf("Could not read group %q after creation: %s", name, err),
		)
		return
	}

	r.populateModelFromResponse(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state adminGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	readResp, err := r.client.GetGroupDetails(name)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateModelFromResponse(readResp, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *adminGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan adminGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	tflog.Debug(ctx, "Updating admin group", map[string]interface{}{
		"name": name,
	})

	params := url.Values{}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		params.Set("description", plan.Description.ValueString())
	}

	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		params.Set("members", plan.Members.ValueString())
	}

	_, err := r.client.SetGroupDetails(name, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Admin Group",
			fmt.Sprintf("Could not update group %q: %s", name, err),
		)
		return
	}

	// Read back.
	readResp, err := r.client.GetGroupDetails(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Admin Group",
			fmt.Sprintf("Could not read group %q after update: %s", name, err),
		)
		return
	}

	r.populateModelFromResponse(readResp, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *adminGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state adminGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting admin group", map[string]interface{}{
		"name": name,
	})

	_, err := r.client.DeleteGroup(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Admin Group",
			fmt.Sprintf("Could not delete group %q: %s", name, err),
		)
	}
}

func (r *adminGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	name := req.ID

	readResp, err := r.client.GetGroupDetails(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing Admin Group",
			fmt.Sprintf("Could not read group %q: %s", name, err),
		)
		return
	}

	model := adminGroupResourceModel{
		Name: types.StringValue(name),
	}

	r.populateModelFromResponse(readResp, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// populateModelFromResponse fills the model fields from an API response map.
func (r *adminGroupResource) populateModelFromResponse(resp map[string]interface{}, model *adminGroupResourceModel) {
	if description, ok := resp["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	if members, ok := resp["members"].([]interface{}); ok {
		names := make([]string, 0, len(members))
		for _, m := range members {
			if mMap, ok := m.(map[string]interface{}); ok {
				if name, ok := mMap["username"].(string); ok {
					names = append(names, name)
				}
			} else if name, ok := m.(string); ok {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			model.Members = types.StringValue(strings.Join(names, ","))
		}
	}
}
