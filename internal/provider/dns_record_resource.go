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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
)

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

type dnsRecordResource struct {
	client *client.Client
}

type dnsRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Zone     types.String `tfsdk:"zone"`
	Domain   types.String `tfsdk:"domain"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Comments types.String `tfsdk:"comments"`
	Priority types.Int64  `tfsdk:"priority"`
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DNS record in a Technitium DNS Server authoritative zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier in the format zone:domain:type:value.",
				Computed:    true,
			},
			"zone": schema.StringAttribute{
				Description: "The authoritative zone name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "The fully qualified domain name of the record.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The DNS record type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"A", "AAAA", "CNAME", "MX", "TXT", "SRV", "NS", "PTR", "CAA", "SOA",
					),
				},
			},
			"value": schema.StringAttribute{
				Description: "The record value. IP address for A/AAAA, hostname for CNAME/NS/PTR/MX, text for TXT, etc.",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Time to live in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the record is disabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"comments": schema.StringAttribute{
				Description: "Comments for the record.",
				Optional:    true,
			},
			"priority": schema.Int64Attribute{
				Description: "Priority value for MX and SRV records.",
				Optional:    true,
			},
		},
	}
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := buildAddParams(&plan)

	tflog.Debug(ctx, "Creating DNS record", map[string]interface{}{
		"zone":   plan.Zone.ValueString(),
		"domain": plan.Domain.ValueString(),
		"type":   plan.Type.ValueString(),
	})

	_, err := r.client.AddRecord(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating DNS Record",
			fmt.Sprintf("Could not create record %s in zone %s: %s",
				plan.Domain.ValueString(), plan.Zone.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(compositeID(
		plan.Zone.ValueString(),
		plan.Domain.ValueString(),
		plan.Type.ValueString(),
		plan.Value.ValueString(),
	))

	r.readRecordIntoModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readRecordIntoModel(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := buildUpdateParams(&state, &plan)

	tflog.Debug(ctx, "Updating DNS record", map[string]interface{}{
		"zone":   plan.Zone.ValueString(),
		"domain": plan.Domain.ValueString(),
		"type":   plan.Type.ValueString(),
	})

	_, err := r.client.UpdateRecord(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating DNS Record",
			fmt.Sprintf("Could not update record %s in zone %s: %s",
				plan.Domain.ValueString(), plan.Zone.ValueString(), err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(compositeID(
		plan.Zone.ValueString(),
		plan.Domain.ValueString(),
		plan.Type.ValueString(),
		plan.Value.ValueString(),
	))

	r.readRecordIntoModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := buildDeleteParams(&state)

	tflog.Debug(ctx, "Deleting DNS record", map[string]interface{}{
		"zone":   state.Zone.ValueString(),
		"domain": state.Domain.ValueString(),
		"type":   state.Type.ValueString(),
	})

	_, err := r.client.DeleteRecord(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DNS Record",
			fmt.Sprintf("Could not delete record %s in zone %s: %s",
				state.Domain.ValueString(), state.Zone.ValueString(), err.Error()),
		)
		return
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 4)
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format zone:domain:type:value, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value"), parts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// readRecordIntoModel fetches the current record from the API and populates
// the model fields. Adds an error diagnostic if the record is not found.
func (r *dnsRecordResource) readRecordIntoModel(_ context.Context, model *dnsRecordResourceModel, diagnostics *diag.Diagnostics) {
	response, err := r.client.GetRecords(
		model.Domain.ValueString(),
		model.Zone.ValueString(),
		false,
	)
	if err != nil {
		diagnostics.AddError(
			"Error Reading DNS Record",
			fmt.Sprintf("Could not read record %s in zone %s: %s",
				model.Domain.ValueString(), model.Zone.ValueString(), err.Error()),
		)
		return
	}

	recordsList, ok := response["records"].([]interface{})
	if !ok {
		diagnostics.AddError(
			"Error Reading DNS Record",
			fmt.Sprintf("Unexpected response format reading record %s in zone %s.",
				model.Domain.ValueString(), model.Zone.ValueString()),
		)
		return
	}

	recType := model.Type.ValueString()
	recValue := model.Value.ValueString()

	var found map[string]interface{}
	for _, item := range recordsList {
		rec, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if stringFromMap(rec, "type") != recType {
			continue
		}
		if recordValueFromRData(rec, recType) == recValue {
			found = rec
			break
		}
	}

	if found == nil {
		diagnostics.AddError(
			"Record Not Found",
			fmt.Sprintf("DNS record %s (type %s, value %s) not found in zone %s.",
				model.Domain.ValueString(), recType, recValue, model.Zone.ValueString()),
		)
		return
	}

	if ttl, ok := found["ttl"].(float64); ok {
		model.TTL = types.Int64Value(int64(ttl))
	}
	model.Disabled = types.BoolValue(boolFromMap(found, "disabled"))

	if comments := stringFromMap(found, "comments"); comments != "" {
		model.Comments = types.StringValue(comments)
	} else if !model.Comments.IsNull() {
		model.Comments = types.StringNull()
	}

	rData, _ := found["rData"].(map[string]interface{})
	if rData != nil && (recType == "MX" || recType == "SRV") {
		var prio float64
		if recType == "MX" {
			prio, _ = rData["preference"].(float64)
		} else {
			prio, _ = rData["priority"].(float64)
		}
		if prio > 0 || !model.Priority.IsNull() {
			model.Priority = types.Int64Value(int64(prio))
		}
	}
}

// recordValueFromRData extracts the primary value string from a record's rData
// map based on the record type.
func recordValueFromRData(rec map[string]interface{}, recordType string) string {
	rData, ok := rec["rData"].(map[string]interface{})
	if !ok {
		return ""
	}
	switch recordType {
	case "A", "AAAA":
		v, _ := rData["ipAddress"].(string)
		return v
	case "CNAME":
		v, _ := rData["cname"].(string)
		return v
	case "NS":
		v, _ := rData["nameServer"].(string)
		return v
	case "PTR":
		v, _ := rData["ptrName"].(string)
		return v
	case "MX":
		v, _ := rData["exchange"].(string)
		return v
	case "TXT":
		v, _ := rData["text"].(string)
		return v
	case "SRV":
		v, _ := rData["target"].(string)
		return v
	case "CAA":
		v, _ := rData["value"].(string)
		return v
	case "SOA":
		v, _ := rData["primaryNameServer"].(string)
		return v
	default:
		return ""
	}
}

// buildAddParams constructs url.Values for the AddRecord API call.
func buildAddParams(plan *dnsRecordResourceModel) url.Values {
	params := url.Values{}
	params.Set("domain", plan.Domain.ValueString())
	params.Set("zone", plan.Zone.ValueString())
	params.Set("type", plan.Type.ValueString())
	params.Set("ttl", fmt.Sprintf("%d", plan.TTL.ValueInt64()))

	if !plan.Disabled.IsNull() && plan.Disabled.ValueBool() {
		params.Set("disabled", "true")
	}

	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		params.Set("comments", plan.Comments.ValueString())
	}

	setValueParams(params, plan.Type.ValueString(), plan.Value.ValueString(), plan.Priority)
	return params
}

// buildUpdateParams constructs url.Values for the UpdateRecord API call.
func buildUpdateParams(state, plan *dnsRecordResourceModel) url.Values {
	params := url.Values{}
	params.Set("domain", plan.Domain.ValueString())
	params.Set("zone", plan.Zone.ValueString())
	params.Set("type", plan.Type.ValueString())
	params.Set("ttl", fmt.Sprintf("%d", plan.TTL.ValueInt64()))

	if !plan.Disabled.IsNull() {
		params.Set("disable", fmt.Sprintf("%t", plan.Disabled.ValueBool()))
	}

	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		params.Set("comments", plan.Comments.ValueString())
	}

	oldValue := state.Value.ValueString()
	newValue := plan.Value.ValueString()
	recType := plan.Type.ValueString()

	switch recType {
	case "A", "AAAA":
		params.Set("ipAddress", oldValue)
		params.Set("newIpAddress", newValue)
	case "CNAME":
		params.Set("cname", newValue)
	case "NS":
		params.Set("nameServer", oldValue)
		params.Set("newNameServer", newValue)
	case "PTR":
		params.Set("ptrName", oldValue)
		params.Set("newPtrName", newValue)
	case "MX":
		params.Set("exchange", oldValue)
		params.Set("newExchange", newValue)
		if !state.Priority.IsNull() && !state.Priority.IsUnknown() {
			params.Set("preference", fmt.Sprintf("%d", state.Priority.ValueInt64()))
		}
		if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
			params.Set("newPreference", fmt.Sprintf("%d", plan.Priority.ValueInt64()))
		}
	case "TXT":
		params.Set("text", oldValue)
		params.Set("newText", newValue)
	case "SRV":
		params.Set("target", oldValue)
		params.Set("newTarget", newValue)
		if !state.Priority.IsNull() && !state.Priority.IsUnknown() {
			params.Set("priority", fmt.Sprintf("%d", state.Priority.ValueInt64()))
		}
		if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
			params.Set("newPriority", fmt.Sprintf("%d", plan.Priority.ValueInt64()))
		}
	case "CAA":
		params.Set("value", oldValue)
		params.Set("newValue", newValue)
	case "SOA":
		params.Set("primaryNameServer", newValue)
	}

	return params
}

// buildDeleteParams constructs url.Values for the DeleteRecord API call.
func buildDeleteParams(state *dnsRecordResourceModel) url.Values {
	params := url.Values{}
	params.Set("domain", state.Domain.ValueString())
	params.Set("zone", state.Zone.ValueString())
	params.Set("type", state.Type.ValueString())

	value := state.Value.ValueString()
	recType := state.Type.ValueString()

	switch recType {
	case "A", "AAAA":
		params.Set("ipAddress", value)
	case "NS":
		params.Set("nameServer", value)
	case "PTR":
		params.Set("ptrName", value)
	case "MX":
		params.Set("exchange", value)
		if !state.Priority.IsNull() && !state.Priority.IsUnknown() {
			params.Set("preference", fmt.Sprintf("%d", state.Priority.ValueInt64()))
		}
	case "TXT":
		params.Set("text", value)
	case "SRV":
		params.Set("target", value)
		if !state.Priority.IsNull() && !state.Priority.IsUnknown() {
			params.Set("priority", fmt.Sprintf("%d", state.Priority.ValueInt64()))
		}
	case "CAA":
		params.Set("value", value)
	}

	return params
}

// setValueParams sets the type-specific value parameters on an Add request.
func setValueParams(params url.Values, recType, value string, priority types.Int64) {
	switch recType {
	case "A", "AAAA":
		params.Set("ipAddress", value)
	case "CNAME":
		params.Set("cname", value)
	case "NS":
		params.Set("nameServer", value)
	case "PTR":
		params.Set("ptrName", value)
	case "MX":
		params.Set("exchange", value)
		if !priority.IsNull() && !priority.IsUnknown() {
			params.Set("preference", fmt.Sprintf("%d", priority.ValueInt64()))
		}
	case "TXT":
		params.Set("text", value)
	case "SRV":
		params.Set("target", value)
		if !priority.IsNull() && !priority.IsUnknown() {
			params.Set("priority", fmt.Sprintf("%d", priority.ValueInt64()))
		}
	case "CAA":
		params.Set("value", value)
	case "SOA":
		params.Set("primaryNameServer", value)
	}
}

// compositeID builds the composite identifier from its components.
func compositeID(zone, domain, recordType, value string) string {
	return fmt.Sprintf("%s:%s:%s:%s", zone, domain, recordType, value)
}
