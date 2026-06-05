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
	ID                types.String `tfsdk:"id"`
	Zone              types.String `tfsdk:"zone"`
	Domain            types.String `tfsdk:"domain"`
	Type              types.String `tfsdk:"type"`
	Value             types.String `tfsdk:"value"`
	TTL               types.Int64  `tfsdk:"ttl"`
	Disabled          types.Bool   `tfsdk:"disabled"`
	Comments          types.String `tfsdk:"comments"`
	Priority          types.Int64  `tfsdk:"priority"`
	Weight            types.Int64  `tfsdk:"weight"`
	Port              types.Int64  `tfsdk:"port"`
	Protocol          types.String `tfsdk:"protocol"`
	Flags             types.Int64  `tfsdk:"flags"`
	Tag               types.String `tfsdk:"tag"`
	ForwarderPriority types.Int64  `tfsdk:"forwarder_priority"`
	DnssecValidation  types.Bool   `tfsdk:"dnssec_validation"`
	ProxyType         types.String `tfsdk:"proxy_type"`
	ProxyAddress      types.String `tfsdk:"proxy_address"`
	ProxyPort         types.Int64  `tfsdk:"proxy_port"`
	ProxyUsername     types.String `tfsdk:"proxy_username"`
	ProxyPassword     types.String `tfsdk:"proxy_password"`
	AppName           types.String `tfsdk:"app_name"`
	ClassPath         types.String `tfsdk:"class_path"`
	RecordData        types.String `tfsdk:"record_data"`
	// ANAME
	AName types.String `tfsdk:"aname"`
	// DNAME
	DName types.String `tfsdk:"dname"`
	// NAPTR
	NaptrOrder       types.Int64  `tfsdk:"naptr_order"`
	NaptrPreference  types.Int64  `tfsdk:"naptr_preference"`
	NaptrFlags       types.String `tfsdk:"naptr_flags"`
	NaptrServices    types.String `tfsdk:"naptr_services"`
	NaptrRegexp      types.String `tfsdk:"naptr_regexp"`
	NaptrReplacement types.String `tfsdk:"naptr_replacement"`
	// SSHFP
	SshfpAlgorithm       types.Int64  `tfsdk:"sshfp_algorithm"`
	SshfpFingerprintType types.Int64  `tfsdk:"sshfp_fingerprint_type"`
	SshfpFingerprint     types.String `tfsdk:"sshfp_fingerprint"`
	// TLSA
	TlsaCertificateUsage           types.String `tfsdk:"tlsa_certificate_usage"`
	TlsaSelector                   types.String `tfsdk:"tlsa_selector"`
	TlsaMatchingType               types.String `tfsdk:"tlsa_matching_type"`
	TlsaCertificateAssociationData types.String `tfsdk:"tlsa_certificate_association_data"`
	// URI
	UriPriority types.Int64  `tfsdk:"uri_priority"`
	UriWeight   types.Int64  `tfsdk:"uri_weight"`
	Uri         types.String `tfsdk:"uri"`
	// DS
	DsKeyTag     types.Int64  `tfsdk:"ds_key_tag"`
	DsAlgorithm  types.Int64  `tfsdk:"ds_algorithm"`
	DsDigestType types.Int64  `tfsdk:"ds_digest_type"`
	DsDigest     types.String `tfsdk:"ds_digest"`
	// SVCB/HTTPS
	SvcPriority   types.Int64  `tfsdk:"svc_priority"`
	SvcTargetName types.String `tfsdk:"svc_target_name"`
	SvcParams     types.String `tfsdk:"svc_params"`
	AutoIpv4Hint  types.Bool   `tfsdk:"auto_ipv4_hint"`
	AutoIpv6Hint  types.Bool   `tfsdk:"auto_ipv6_hint"`
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
						"A", "AAAA", "CNAME", "MX", "TXT", "SRV", "NS", "PTR", "CAA", "SOA", "FWD", "APP",
						"ANAME", "DNAME", "NAPTR", "SSHFP", "TLSA", "URI", "DS", "SVCB", "HTTPS",
					),
				},
			},
			"value": schema.StringAttribute{
				Description: "The record value. IP address for A/AAAA, hostname for CNAME/NS/PTR/MX/SRV, text for TXT, forwarder address for FWD.",
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
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Weight value for SRV records.",
				Optional:    true,
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number for SRV records.",
				Optional:    true,
				Computed:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "Forwarding protocol for FWD records. Valid values: Udp, Tcp, Tls, Https, Quic.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Udp", "Tcp", "Tls", "Https", "Quic"),
				},
			},
			"flags": schema.Int64Attribute{
				Description: "Flags value for CAA records.",
				Optional:    true,
				Computed:    true,
			},
			"tag": schema.StringAttribute{
				Description: "Tag value for CAA records (e.g. issue, issuewild, iodef).",
				Optional:    true,
				Computed:    true,
			},
			"forwarder_priority": schema.Int64Attribute{
				Description: "Priority for FWD records. Lower values are queried first; equal values are queried concurrently.",
				Optional:    true,
				Computed:    true,
			},
			"dnssec_validation": schema.BoolAttribute{
				Description: "Enable DNSSEC validation for FWD records.",
				Optional:    true,
			},
			"proxy_type": schema.StringAttribute{
				Description: "Proxy type for FWD records. Valid values: NoProxy, DefaultProxy, Http, Socks5.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NoProxy", "DefaultProxy", "Http", "Socks5"),
				},
			},
			"proxy_address": schema.StringAttribute{
				Description: "Proxy server address for FWD records.",
				Optional:    true,
			},
			"proxy_port": schema.Int64Attribute{
				Description: "Proxy server port for FWD records.",
				Optional:    true,
			},
			"proxy_username": schema.StringAttribute{
				Description: "Proxy server username for FWD records.",
				Optional:    true,
				Sensitive:   true,
			},
			"proxy_password": schema.StringAttribute{
				Description: "Proxy server password for FWD records.",
				Optional:    true,
				Sensitive:   true,
			},
			"app_name": schema.StringAttribute{
				Description: "DNS app name for APP records.",
				Optional:    true,
			},
			"class_path": schema.StringAttribute{
				Description: "Class path for APP records.",
				Optional:    true,
			},
			"record_data": schema.StringAttribute{
				Description: "App-specific record data for APP records.",
				Optional:    true,
			},
			// ANAME
			"aname": schema.StringAttribute{
				Description: "Target domain for ANAME records.",
				Optional:    true,
			},
			// DNAME
			"dname": schema.StringAttribute{
				Description: "Target domain for DNAME records.",
				Optional:    true,
			},
			// NAPTR
			"naptr_order": schema.Int64Attribute{
				Description: "Order value for NAPTR records.",
				Optional:    true,
			},
			"naptr_preference": schema.Int64Attribute{
				Description: "Preference value for NAPTR records.",
				Optional:    true,
			},
			"naptr_flags": schema.StringAttribute{
				Description: "Flags for NAPTR records.",
				Optional:    true,
			},
			"naptr_services": schema.StringAttribute{
				Description: "Services field for NAPTR records.",
				Optional:    true,
			},
			"naptr_regexp": schema.StringAttribute{
				Description: "Regular expression for NAPTR records.",
				Optional:    true,
			},
			"naptr_replacement": schema.StringAttribute{
				Description: "Replacement domain for NAPTR records.",
				Optional:    true,
			},
			// SSHFP
			"sshfp_algorithm": schema.Int64Attribute{
				Description: "Algorithm number for SSHFP records (1=RSA, 2=DSA, 3=ECDSA, 4=Ed25519).",
				Optional:    true,
			},
			"sshfp_fingerprint_type": schema.Int64Attribute{
				Description: "Fingerprint type for SSHFP records (1=SHA-1, 2=SHA-256).",
				Optional:    true,
			},
			"sshfp_fingerprint": schema.StringAttribute{
				Description: "Fingerprint hex string for SSHFP records.",
				Optional:    true,
			},
			// TLSA
			"tlsa_certificate_usage": schema.StringAttribute{
				Description: "Certificate usage for TLSA records.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("PKIX-TA", "PKIX-EE", "DANE-TA", "DANE-EE"),
				},
			},
			"tlsa_selector": schema.StringAttribute{
				Description: "Selector for TLSA records.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Cert", "SPKI"),
				},
			},
			"tlsa_matching_type": schema.StringAttribute{
				Description: "Matching type for TLSA records.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Full", "SHA2-256", "SHA2-512"),
				},
			},
			"tlsa_certificate_association_data": schema.StringAttribute{
				Description: "Certificate association data for TLSA records.",
				Optional:    true,
			},
			// URI
			"uri_priority": schema.Int64Attribute{
				Description: "Priority for URI records.",
				Optional:    true,
			},
			"uri_weight": schema.Int64Attribute{
				Description: "Weight for URI records.",
				Optional:    true,
			},
			"uri": schema.StringAttribute{
				Description: "URI value for URI records.",
				Optional:    true,
			},
			// DS
			"ds_key_tag": schema.Int64Attribute{
				Description: "Key tag for DS records.",
				Optional:    true,
			},
			"ds_algorithm": schema.Int64Attribute{
				Description: "Algorithm number for DS records.",
				Optional:    true,
			},
			"ds_digest_type": schema.Int64Attribute{
				Description: "Digest type for DS records.",
				Optional:    true,
			},
			"ds_digest": schema.StringAttribute{
				Description: "Digest hex string for DS records.",
				Optional:    true,
			},
			// SVCB/HTTPS
			"svc_priority": schema.Int64Attribute{
				Description: "Priority for SVCB/HTTPS records.",
				Optional:    true,
			},
			"svc_target_name": schema.StringAttribute{
				Description: "Target name for SVCB/HTTPS records.",
				Optional:    true,
			},
			"svc_params": schema.StringAttribute{
				Description: "Service parameters for SVCB/HTTPS records.",
				Optional:    true,
			},
			"auto_ipv4_hint": schema.BoolAttribute{
				Description: "Automatically generate IPv4 address hints for SVCB/HTTPS records.",
				Optional:    true,
			},
			"auto_ipv6_hint": schema.BoolAttribute{
				Description: "Automatically generate IPv6 address hints for SVCB/HTTPS records.",
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

	if !plan.Disabled.IsNull() && plan.Disabled.ValueBool() {
		disableParams := url.Values{}
		disableParams.Set("domain", plan.Domain.ValueString())
		disableParams.Set("zone", plan.Zone.ValueString())
		disableParams.Set("type", plan.Type.ValueString())
		disableParams.Set("disable", "true")
		setValueParams(disableParams, &plan)
		_, err = r.client.UpdateRecord(disableParams)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Disabling DNS Record",
				fmt.Sprintf("Record was created but could not be disabled: %s", err.Error()),
			)
			return
		}
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

	var readDiags diag.Diagnostics
	r.readRecordIntoModel(ctx, &state, &readDiags)
	if readDiags.HasError() {
		for _, d := range readDiags {
			if d.Summary() == "Record Not Found" {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.Append(readDiags...)
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
		if strings.Contains(err.Error(), "no such record") ||
			strings.Contains(err.Error(), "was not found") ||
			strings.Contains(err.Error(), "No such zone") {
			return
		}
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
		if strings.EqualFold(recordValueFromRData(rec, recType), recValue) {
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

	apiValue := recordValueFromRData(found, recType)
	if apiValue != "" && !strings.EqualFold(apiValue, model.Value.ValueString()) {
		model.Value = types.StringValue(apiValue)
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

	nullifyUnknowns(model)

	rData, _ := found["rData"].(map[string]interface{})
	if rData != nil {
		switch recType {
		case "MX":
			if pref, ok := rData["preference"].(float64); ok {
				model.Priority = types.Int64Value(int64(pref))
			}
		case "SRV":
			if prio, ok := rData["priority"].(float64); ok {
				model.Priority = types.Int64Value(int64(prio))
			}
			if w, ok := rData["weight"].(float64); ok {
				model.Weight = types.Int64Value(int64(w))
			}
			if p, ok := rData["port"].(float64); ok {
				model.Port = types.Int64Value(int64(p))
			}
		case "CAA":
			if f, ok := rData["flags"].(float64); ok {
				model.Flags = types.Int64Value(int64(f))
			}
			if tag, ok := rData["tag"].(string); ok && tag != "" {
				model.Tag = types.StringValue(tag)
			} else if !model.Tag.IsNull() {
				model.Tag = types.StringNull()
			}
		case "FWD":
			if proto, ok := rData["protocol"].(string); ok && proto != "" {
				model.Protocol = types.StringValue(proto)
			}
			if fp, ok := rData["forwarderPriority"].(float64); ok {
				model.ForwarderPriority = types.Int64Value(int64(fp))
			}
			if dv, ok := rData["dnssecValidation"].(bool); ok && (dv || !model.DnssecValidation.IsNull()) {
				model.DnssecValidation = types.BoolValue(dv)
			}
			if pt, ok := rData["proxyType"].(string); ok && pt != "" && (pt != "DefaultProxy" || !model.ProxyType.IsNull()) {
				model.ProxyType = types.StringValue(pt)
			}
			if pa, ok := rData["proxyAddress"].(string); ok && pa != "" {
				model.ProxyAddress = types.StringValue(pa)
			} else if !model.ProxyAddress.IsNull() {
				model.ProxyAddress = types.StringNull()
			}
			if pp, ok := rData["proxyPort"].(float64); ok && (pp > 0 || !model.ProxyPort.IsNull()) {
				model.ProxyPort = types.Int64Value(int64(pp))
			}
			if pu, ok := rData["proxyUsername"].(string); ok && pu != "" {
				model.ProxyUsername = types.StringValue(pu)
			} else if !model.ProxyUsername.IsNull() {
				model.ProxyUsername = types.StringNull()
			}
			if pp, ok := rData["proxyPassword"].(string); ok && pp != "" {
				model.ProxyPassword = types.StringValue(pp)
			} else if !model.ProxyPassword.IsNull() {
				model.ProxyPassword = types.StringNull()
			}
		case "APP":
			if an, ok := rData["appName"].(string); ok && an != "" {
				model.AppName = types.StringValue(an)
			}
			if cp, ok := rData["classPath"].(string); ok && cp != "" {
				model.ClassPath = types.StringValue(cp)
			}
			if rd, ok := rData["recordData"].(string); ok && rd != "" {
				model.RecordData = types.StringValue(rd)
			} else if !model.RecordData.IsNull() {
				model.RecordData = types.StringNull()
			}
		case "ANAME":
			if v, ok := rData["aname"].(string); ok && v != "" {
				model.AName = types.StringValue(v)
			} else if !model.AName.IsNull() {
				model.AName = types.StringNull()
			}
		case "DNAME":
			if v, ok := rData["dname"].(string); ok && v != "" {
				model.DName = types.StringValue(v)
			} else if !model.DName.IsNull() {
				model.DName = types.StringNull()
			}
		case "NAPTR":
			if v, ok := rData["order"].(float64); ok {
				model.NaptrOrder = types.Int64Value(int64(v))
			}
			if v, ok := rData["preference"].(float64); ok {
				model.NaptrPreference = types.Int64Value(int64(v))
			}
			if v, ok := rData["flags"].(string); ok && v != "" {
				model.NaptrFlags = types.StringValue(v)
			} else if !model.NaptrFlags.IsNull() {
				model.NaptrFlags = types.StringNull()
			}
			if v, ok := rData["services"].(string); ok && v != "" {
				model.NaptrServices = types.StringValue(v)
			} else if !model.NaptrServices.IsNull() {
				model.NaptrServices = types.StringNull()
			}
			if v, ok := rData["regexp"].(string); ok && v != "" {
				model.NaptrRegexp = types.StringValue(v)
			} else if !model.NaptrRegexp.IsNull() {
				model.NaptrRegexp = types.StringNull()
			}
			if v, ok := rData["replacement"].(string); ok && v != "" {
				model.NaptrReplacement = types.StringValue(v)
			} else if !model.NaptrReplacement.IsNull() {
				model.NaptrReplacement = types.StringNull()
			}
		case "SSHFP":
			if v, ok := rData["algorithm"].(string); ok && v != "" {
				model.SshfpAlgorithm = types.Int64Value(sshfpAlgorithmToInt(v))
			} else if v, ok := rData["sshfpAlgorithm"].(float64); ok {
				model.SshfpAlgorithm = types.Int64Value(int64(v))
			}
			if v, ok := rData["fingerprintType"].(string); ok && v != "" {
				model.SshfpFingerprintType = types.Int64Value(sshfpFingerprintTypeToInt(v))
			} else if v, ok := rData["sshfpFingerprintType"].(float64); ok {
				model.SshfpFingerprintType = types.Int64Value(int64(v))
			}
			if v, ok := rData["fingerprint"].(string); ok && v != "" {
				if !strings.EqualFold(v, model.SshfpFingerprint.ValueString()) {
					model.SshfpFingerprint = types.StringValue(v)
				}
			} else if !model.SshfpFingerprint.IsNull() {
				model.SshfpFingerprint = types.StringNull()
			}
		case "TLSA":
			if v, ok := rData["certificateUsage"].(string); ok && v != "" {
				model.TlsaCertificateUsage = types.StringValue(v)
			} else if !model.TlsaCertificateUsage.IsNull() {
				model.TlsaCertificateUsage = types.StringNull()
			}
			if v, ok := rData["selector"].(string); ok && v != "" {
				model.TlsaSelector = types.StringValue(v)
			} else if !model.TlsaSelector.IsNull() {
				model.TlsaSelector = types.StringNull()
			}
			if v, ok := rData["matchingType"].(string); ok && v != "" {
				model.TlsaMatchingType = types.StringValue(v)
			} else if !model.TlsaMatchingType.IsNull() {
				model.TlsaMatchingType = types.StringNull()
			}
			if v, ok := rData["certificateAssociationData"].(string); ok && v != "" {
				if !strings.EqualFold(v, model.TlsaCertificateAssociationData.ValueString()) {
					model.TlsaCertificateAssociationData = types.StringValue(v)
				}
			} else if !model.TlsaCertificateAssociationData.IsNull() {
				model.TlsaCertificateAssociationData = types.StringNull()
			}
		case "URI":
			if v, ok := rData["priority"].(float64); ok {
				model.UriPriority = types.Int64Value(int64(v))
			}
			if v, ok := rData["weight"].(float64); ok {
				model.UriWeight = types.Int64Value(int64(v))
			}
			if v, ok := rData["uri"].(string); ok && v != "" {
				model.Uri = types.StringValue(v)
			} else if !model.Uri.IsNull() {
				model.Uri = types.StringNull()
			}
		case "DS":
			if v, ok := rData["keyTag"].(float64); ok {
				model.DsKeyTag = types.Int64Value(int64(v))
			}
			if v, ok := rData["algorithmNumber"].(float64); ok {
				model.DsAlgorithm = types.Int64Value(int64(v))
			} else if v, ok := rData["algorithm"].(float64); ok {
				model.DsAlgorithm = types.Int64Value(int64(v))
			}
			if v, ok := rData["digestTypeNumber"].(float64); ok {
				model.DsDigestType = types.Int64Value(int64(v))
			} else if v, ok := rData["digestType"].(float64); ok {
				model.DsDigestType = types.Int64Value(int64(v))
			}
			if v, ok := rData["digest"].(string); ok && v != "" {
				if !strings.EqualFold(v, model.DsDigest.ValueString()) {
					model.DsDigest = types.StringValue(v)
				}
			} else if !model.DsDigest.IsNull() {
				model.DsDigest = types.StringNull()
			}
		case "SVCB", "HTTPS":
			if v, ok := rData["svcPriority"].(float64); ok {
				model.SvcPriority = types.Int64Value(int64(v))
			}
			if v, ok := rData["svcTargetName"].(string); ok && v != "" {
				model.SvcTargetName = types.StringValue(v)
			} else if !model.SvcTargetName.IsNull() {
				model.SvcTargetName = types.StringNull()
			}
			if v, ok := rData["svcParams"].(map[string]interface{}); ok && len(v) > 0 {
				var parts []string
				for key, val := range v {
					if s, ok := val.(string); ok {
						parts = append(parts, key, s)
					}
				}
				if len(parts) > 0 {
					model.SvcParams = types.StringValue(strings.Join(parts, "|"))
				}
			} else if sv, ok := rData["svcParams"].(string); ok && sv != "" {
				model.SvcParams = types.StringValue(sv)
			} else if !model.SvcParams.IsNull() {
				model.SvcParams = types.StringNull()
			}
			if v, ok := rData["autoIpv4Hint"].(bool); ok && (v || !model.AutoIpv4Hint.IsNull()) {
				model.AutoIpv4Hint = types.BoolValue(v)
			}
			if v, ok := rData["autoIpv6Hint"].(bool); ok && (v || !model.AutoIpv6Hint.IsNull()) {
				model.AutoIpv6Hint = types.BoolValue(v)
			}
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
	case "FWD":
		v, _ := rData["forwarder"].(string)
		return v
	case "APP":
		v, _ := rData["appName"].(string)
		return v
	case "ANAME":
		v, _ := rData["aname"].(string)
		return v
	case "DNAME":
		v, _ := rData["dname"].(string)
		return v
	case "NAPTR":
		v, _ := rData["replacement"].(string)
		return v
	case "SSHFP":
		v, _ := rData["fingerprint"].(string)
		return v
	case "TLSA":
		v, _ := rData["certificateAssociationData"].(string)
		return v
	case "URI":
		v, _ := rData["uri"].(string)
		return v
	case "DS":
		v, _ := rData["digest"].(string)
		return v
	case "SVCB", "HTTPS":
		v, _ := rData["svcTargetName"].(string)
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

	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		params.Set("comments", plan.Comments.ValueString())
	}

	setValueParams(params, plan)
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
		} else {
			params.Set("priority", "0")
		}
		if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
			params.Set("newPriority", fmt.Sprintf("%d", plan.Priority.ValueInt64()))
		} else {
			params.Set("newPriority", "0")
		}
		if !state.Weight.IsNull() && !state.Weight.IsUnknown() {
			params.Set("weight", fmt.Sprintf("%d", state.Weight.ValueInt64()))
		} else {
			params.Set("weight", "0")
		}
		if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
			params.Set("newWeight", fmt.Sprintf("%d", plan.Weight.ValueInt64()))
		} else {
			params.Set("newWeight", "0")
		}
		if !state.Port.IsNull() && !state.Port.IsUnknown() {
			params.Set("port", fmt.Sprintf("%d", state.Port.ValueInt64()))
		} else {
			params.Set("port", "0")
		}
		if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
			params.Set("newPort", fmt.Sprintf("%d", plan.Port.ValueInt64()))
		} else {
			params.Set("newPort", "0")
		}
	case "CAA":
		params.Set("value", oldValue)
		params.Set("newValue", newValue)
		if !state.Flags.IsNull() && !state.Flags.IsUnknown() {
			params.Set("flags", fmt.Sprintf("%d", state.Flags.ValueInt64()))
		}
		if !plan.Flags.IsNull() && !plan.Flags.IsUnknown() {
			params.Set("newFlags", fmt.Sprintf("%d", plan.Flags.ValueInt64()))
		}
		if !state.Tag.IsNull() && !state.Tag.IsUnknown() {
			params.Set("tag", state.Tag.ValueString())
		}
		if !plan.Tag.IsNull() && !plan.Tag.IsUnknown() {
			params.Set("newTag", plan.Tag.ValueString())
		}
	case "SOA":
		params.Set("primaryNameServer", newValue)
	case "FWD":
		params.Set("forwarder", oldValue)
		params.Set("newForwarder", newValue)
		if !state.Protocol.IsNull() && !state.Protocol.IsUnknown() {
			params.Set("protocol", state.Protocol.ValueString())
		} else {
			params.Set("protocol", "Udp")
		}
		if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() {
			params.Set("newProtocol", plan.Protocol.ValueString())
		} else {
			params.Set("newProtocol", "Udp")
		}
		if !plan.ForwarderPriority.IsNull() && !plan.ForwarderPriority.IsUnknown() {
			params.Set("forwarderPriority", fmt.Sprintf("%d", plan.ForwarderPriority.ValueInt64()))
		}
		if !plan.DnssecValidation.IsNull() && !plan.DnssecValidation.IsUnknown() {
			params.Set("dnssecValidation", fmt.Sprintf("%t", plan.DnssecValidation.ValueBool()))
		}
		if !plan.ProxyType.IsNull() && !plan.ProxyType.IsUnknown() {
			params.Set("proxyType", plan.ProxyType.ValueString())
		}
		if !plan.ProxyAddress.IsNull() && !plan.ProxyAddress.IsUnknown() {
			params.Set("proxyAddress", plan.ProxyAddress.ValueString())
		}
		if !plan.ProxyPort.IsNull() && !plan.ProxyPort.IsUnknown() {
			params.Set("proxyPort", fmt.Sprintf("%d", plan.ProxyPort.ValueInt64()))
		}
		if !plan.ProxyUsername.IsNull() && !plan.ProxyUsername.IsUnknown() {
			params.Set("proxyUsername", plan.ProxyUsername.ValueString())
		}
		if !plan.ProxyPassword.IsNull() && !plan.ProxyPassword.IsUnknown() {
			params.Set("proxyPassword", plan.ProxyPassword.ValueString())
		}
	case "APP":
		if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
			params.Set("appName", plan.AppName.ValueString())
		}
		if !plan.ClassPath.IsNull() && !plan.ClassPath.IsUnknown() {
			params.Set("classPath", plan.ClassPath.ValueString())
		}
		if !plan.RecordData.IsNull() && !plan.RecordData.IsUnknown() {
			params.Set("recordData", plan.RecordData.ValueString())
		}
	case "ANAME":
		params.Set("aname", oldValue)
		params.Set("newAName", newValue)
	case "DNAME":
		params.Set("dname", newValue)
	case "NAPTR":
		if !state.NaptrOrder.IsNull() && !state.NaptrOrder.IsUnknown() {
			params.Set("naptrOrder", fmt.Sprintf("%d", state.NaptrOrder.ValueInt64()))
		}
		if !plan.NaptrOrder.IsNull() && !plan.NaptrOrder.IsUnknown() {
			params.Set("naptrNewOrder", fmt.Sprintf("%d", plan.NaptrOrder.ValueInt64()))
		}
		if !state.NaptrPreference.IsNull() && !state.NaptrPreference.IsUnknown() {
			params.Set("naptrPreference", fmt.Sprintf("%d", state.NaptrPreference.ValueInt64()))
		}
		if !plan.NaptrPreference.IsNull() && !plan.NaptrPreference.IsUnknown() {
			params.Set("naptrNewPreference", fmt.Sprintf("%d", plan.NaptrPreference.ValueInt64()))
		}
		if !state.NaptrFlags.IsNull() && !state.NaptrFlags.IsUnknown() {
			params.Set("naptrFlags", state.NaptrFlags.ValueString())
		}
		if !plan.NaptrFlags.IsNull() && !plan.NaptrFlags.IsUnknown() {
			params.Set("naptrNewFlags", plan.NaptrFlags.ValueString())
		}
		if !state.NaptrServices.IsNull() && !state.NaptrServices.IsUnknown() {
			params.Set("naptrServices", state.NaptrServices.ValueString())
		}
		if !plan.NaptrServices.IsNull() && !plan.NaptrServices.IsUnknown() {
			params.Set("naptrNewServices", plan.NaptrServices.ValueString())
		}
		if !state.NaptrRegexp.IsNull() && !state.NaptrRegexp.IsUnknown() {
			params.Set("naptrRegexp", state.NaptrRegexp.ValueString())
		}
		if !plan.NaptrRegexp.IsNull() && !plan.NaptrRegexp.IsUnknown() {
			params.Set("naptrNewRegexp", plan.NaptrRegexp.ValueString())
		}
		if !state.NaptrReplacement.IsNull() && !state.NaptrReplacement.IsUnknown() {
			params.Set("naptrReplacement", state.NaptrReplacement.ValueString())
		}
		if !plan.NaptrReplacement.IsNull() && !plan.NaptrReplacement.IsUnknown() {
			params.Set("naptrNewReplacement", plan.NaptrReplacement.ValueString())
		}
	case "SSHFP":
		if !state.SshfpAlgorithm.IsNull() && !state.SshfpAlgorithm.IsUnknown() {
			params.Set("sshfpAlgorithm", fmt.Sprintf("%d", state.SshfpAlgorithm.ValueInt64()))
		}
		if !plan.SshfpAlgorithm.IsNull() && !plan.SshfpAlgorithm.IsUnknown() {
			params.Set("newSshfpAlgorithm", fmt.Sprintf("%d", plan.SshfpAlgorithm.ValueInt64()))
		}
		if !state.SshfpFingerprintType.IsNull() && !state.SshfpFingerprintType.IsUnknown() {
			params.Set("sshfpFingerprintType", fmt.Sprintf("%d", state.SshfpFingerprintType.ValueInt64()))
		}
		if !plan.SshfpFingerprintType.IsNull() && !plan.SshfpFingerprintType.IsUnknown() {
			params.Set("newSshfpFingerprintType", fmt.Sprintf("%d", plan.SshfpFingerprintType.ValueInt64()))
		}
		if !state.SshfpFingerprint.IsNull() && !state.SshfpFingerprint.IsUnknown() {
			params.Set("sshfpFingerprint", state.SshfpFingerprint.ValueString())
		}
		if !plan.SshfpFingerprint.IsNull() && !plan.SshfpFingerprint.IsUnknown() {
			params.Set("newSshfpFingerprint", plan.SshfpFingerprint.ValueString())
		}
	case "TLSA":
		if !state.TlsaCertificateUsage.IsNull() && !state.TlsaCertificateUsage.IsUnknown() {
			params.Set("tlsaCertificateUsage", state.TlsaCertificateUsage.ValueString())
		}
		if !plan.TlsaCertificateUsage.IsNull() && !plan.TlsaCertificateUsage.IsUnknown() {
			params.Set("newTlsaCertificateUsage", plan.TlsaCertificateUsage.ValueString())
		}
		if !state.TlsaSelector.IsNull() && !state.TlsaSelector.IsUnknown() {
			params.Set("tlsaSelector", state.TlsaSelector.ValueString())
		}
		if !plan.TlsaSelector.IsNull() && !plan.TlsaSelector.IsUnknown() {
			params.Set("newTlsaSelector", plan.TlsaSelector.ValueString())
		}
		if !state.TlsaMatchingType.IsNull() && !state.TlsaMatchingType.IsUnknown() {
			params.Set("tlsaMatchingType", state.TlsaMatchingType.ValueString())
		}
		if !plan.TlsaMatchingType.IsNull() && !plan.TlsaMatchingType.IsUnknown() {
			params.Set("newTlsaMatchingType", plan.TlsaMatchingType.ValueString())
		}
		if !state.TlsaCertificateAssociationData.IsNull() && !state.TlsaCertificateAssociationData.IsUnknown() {
			params.Set("tlsaCertificateAssociationData", state.TlsaCertificateAssociationData.ValueString())
		}
		if !plan.TlsaCertificateAssociationData.IsNull() && !plan.TlsaCertificateAssociationData.IsUnknown() {
			params.Set("newTlsaCertificateAssociationData", plan.TlsaCertificateAssociationData.ValueString())
		}
	case "URI":
		if !state.UriPriority.IsNull() && !state.UriPriority.IsUnknown() {
			params.Set("uriPriority", fmt.Sprintf("%d", state.UriPriority.ValueInt64()))
		}
		if !plan.UriPriority.IsNull() && !plan.UriPriority.IsUnknown() {
			params.Set("newUriPriority", fmt.Sprintf("%d", plan.UriPriority.ValueInt64()))
		}
		if !state.UriWeight.IsNull() && !state.UriWeight.IsUnknown() {
			params.Set("uriWeight", fmt.Sprintf("%d", state.UriWeight.ValueInt64()))
		}
		if !plan.UriWeight.IsNull() && !plan.UriWeight.IsUnknown() {
			params.Set("newUriWeight", fmt.Sprintf("%d", plan.UriWeight.ValueInt64()))
		}
		if !state.Uri.IsNull() && !state.Uri.IsUnknown() {
			params.Set("uri", state.Uri.ValueString())
		}
		if !plan.Uri.IsNull() && !plan.Uri.IsUnknown() {
			params.Set("newUri", plan.Uri.ValueString())
		}
	case "DS":
		if !state.DsKeyTag.IsNull() && !state.DsKeyTag.IsUnknown() {
			params.Set("keyTag", fmt.Sprintf("%d", state.DsKeyTag.ValueInt64()))
		}
		if !plan.DsKeyTag.IsNull() && !plan.DsKeyTag.IsUnknown() {
			params.Set("newKeyTag", fmt.Sprintf("%d", plan.DsKeyTag.ValueInt64()))
		}
		if !state.DsAlgorithm.IsNull() && !state.DsAlgorithm.IsUnknown() {
			params.Set("algorithm", fmt.Sprintf("%d", state.DsAlgorithm.ValueInt64()))
		}
		if !plan.DsAlgorithm.IsNull() && !plan.DsAlgorithm.IsUnknown() {
			params.Set("newAlgorithm", fmt.Sprintf("%d", plan.DsAlgorithm.ValueInt64()))
		}
		if !state.DsDigestType.IsNull() && !state.DsDigestType.IsUnknown() {
			params.Set("digestType", fmt.Sprintf("%d", state.DsDigestType.ValueInt64()))
		}
		if !plan.DsDigestType.IsNull() && !plan.DsDigestType.IsUnknown() {
			params.Set("newDigestType", fmt.Sprintf("%d", plan.DsDigestType.ValueInt64()))
		}
		if !state.DsDigest.IsNull() && !state.DsDigest.IsUnknown() {
			params.Set("digest", state.DsDigest.ValueString())
		}
		if !plan.DsDigest.IsNull() && !plan.DsDigest.IsUnknown() {
			params.Set("newDigest", plan.DsDigest.ValueString())
		}
	case "SVCB", "HTTPS":
		if !state.SvcPriority.IsNull() && !state.SvcPriority.IsUnknown() {
			params.Set("svcPriority", fmt.Sprintf("%d", state.SvcPriority.ValueInt64()))
		}
		if !plan.SvcPriority.IsNull() && !plan.SvcPriority.IsUnknown() {
			params.Set("newSvcPriority", fmt.Sprintf("%d", plan.SvcPriority.ValueInt64()))
		}
		if !state.SvcTargetName.IsNull() && !state.SvcTargetName.IsUnknown() {
			params.Set("svcTargetName", state.SvcTargetName.ValueString())
		}
		if !plan.SvcTargetName.IsNull() && !plan.SvcTargetName.IsUnknown() {
			params.Set("newSvcTargetName", plan.SvcTargetName.ValueString())
		}
		if !state.SvcParams.IsNull() && !state.SvcParams.IsUnknown() {
			params.Set("svcParams", state.SvcParams.ValueString())
		}
		if !plan.SvcParams.IsNull() && !plan.SvcParams.IsUnknown() {
			params.Set("newSvcParams", plan.SvcParams.ValueString())
		}
		if !plan.AutoIpv4Hint.IsNull() && !plan.AutoIpv4Hint.IsUnknown() {
			params.Set("autoIpv4Hint", fmt.Sprintf("%t", plan.AutoIpv4Hint.ValueBool()))
		}
		if !plan.AutoIpv6Hint.IsNull() && !plan.AutoIpv6Hint.IsUnknown() {
			params.Set("autoIpv6Hint", fmt.Sprintf("%t", plan.AutoIpv6Hint.ValueBool()))
		}
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
		if !state.Weight.IsNull() && !state.Weight.IsUnknown() {
			params.Set("weight", fmt.Sprintf("%d", state.Weight.ValueInt64()))
		}
		if !state.Port.IsNull() && !state.Port.IsUnknown() {
			params.Set("port", fmt.Sprintf("%d", state.Port.ValueInt64()))
		}
	case "CAA":
		params.Set("value", value)
		if !state.Flags.IsNull() && !state.Flags.IsUnknown() {
			params.Set("flags", fmt.Sprintf("%d", state.Flags.ValueInt64()))
		}
		if !state.Tag.IsNull() && !state.Tag.IsUnknown() {
			params.Set("tag", state.Tag.ValueString())
		}
	case "FWD":
		params.Set("forwarder", value)
		if !state.Protocol.IsNull() && !state.Protocol.IsUnknown() {
			params.Set("protocol", state.Protocol.ValueString())
		}
	case "APP":
		if !state.AppName.IsNull() && !state.AppName.IsUnknown() {
			params.Set("appName", state.AppName.ValueString())
		}
		if !state.ClassPath.IsNull() && !state.ClassPath.IsUnknown() {
			params.Set("classPath", state.ClassPath.ValueString())
		}
		if !state.RecordData.IsNull() && !state.RecordData.IsUnknown() {
			params.Set("recordData", state.RecordData.ValueString())
		}
	case "CNAME":
		params.Set("cname", value)
	case "ANAME":
		params.Set("aname", value)
	case "DNAME":
		params.Set("dname", value)
	case "NAPTR":
		if !state.NaptrOrder.IsNull() && !state.NaptrOrder.IsUnknown() {
			params.Set("naptrOrder", fmt.Sprintf("%d", state.NaptrOrder.ValueInt64()))
		}
		if !state.NaptrPreference.IsNull() && !state.NaptrPreference.IsUnknown() {
			params.Set("naptrPreference", fmt.Sprintf("%d", state.NaptrPreference.ValueInt64()))
		}
		if !state.NaptrFlags.IsNull() && !state.NaptrFlags.IsUnknown() {
			params.Set("naptrFlags", state.NaptrFlags.ValueString())
		}
		if !state.NaptrServices.IsNull() && !state.NaptrServices.IsUnknown() {
			params.Set("naptrServices", state.NaptrServices.ValueString())
		}
		if !state.NaptrRegexp.IsNull() && !state.NaptrRegexp.IsUnknown() {
			params.Set("naptrRegexp", state.NaptrRegexp.ValueString())
		}
		if !state.NaptrReplacement.IsNull() && !state.NaptrReplacement.IsUnknown() {
			params.Set("naptrReplacement", state.NaptrReplacement.ValueString())
		}
	case "SSHFP":
		if !state.SshfpAlgorithm.IsNull() && !state.SshfpAlgorithm.IsUnknown() {
			params.Set("sshfpAlgorithm", fmt.Sprintf("%d", state.SshfpAlgorithm.ValueInt64()))
		}
		if !state.SshfpFingerprintType.IsNull() && !state.SshfpFingerprintType.IsUnknown() {
			params.Set("sshfpFingerprintType", fmt.Sprintf("%d", state.SshfpFingerprintType.ValueInt64()))
		}
		if !state.SshfpFingerprint.IsNull() && !state.SshfpFingerprint.IsUnknown() {
			params.Set("sshfpFingerprint", state.SshfpFingerprint.ValueString())
		}
	case "TLSA":
		if !state.TlsaCertificateUsage.IsNull() && !state.TlsaCertificateUsage.IsUnknown() {
			params.Set("tlsaCertificateUsage", state.TlsaCertificateUsage.ValueString())
		}
		if !state.TlsaSelector.IsNull() && !state.TlsaSelector.IsUnknown() {
			params.Set("tlsaSelector", state.TlsaSelector.ValueString())
		}
		if !state.TlsaMatchingType.IsNull() && !state.TlsaMatchingType.IsUnknown() {
			params.Set("tlsaMatchingType", state.TlsaMatchingType.ValueString())
		}
		if !state.TlsaCertificateAssociationData.IsNull() && !state.TlsaCertificateAssociationData.IsUnknown() {
			params.Set("tlsaCertificateAssociationData", state.TlsaCertificateAssociationData.ValueString())
		}
	case "URI":
		if !state.UriPriority.IsNull() && !state.UriPriority.IsUnknown() {
			params.Set("uriPriority", fmt.Sprintf("%d", state.UriPriority.ValueInt64()))
		}
		if !state.UriWeight.IsNull() && !state.UriWeight.IsUnknown() {
			params.Set("uriWeight", fmt.Sprintf("%d", state.UriWeight.ValueInt64()))
		}
		if !state.Uri.IsNull() && !state.Uri.IsUnknown() {
			params.Set("uri", state.Uri.ValueString())
		}
	case "DS":
		if !state.DsKeyTag.IsNull() && !state.DsKeyTag.IsUnknown() {
			params.Set("keyTag", fmt.Sprintf("%d", state.DsKeyTag.ValueInt64()))
		}
		if !state.DsAlgorithm.IsNull() && !state.DsAlgorithm.IsUnknown() {
			params.Set("algorithm", fmt.Sprintf("%d", state.DsAlgorithm.ValueInt64()))
		}
		if !state.DsDigestType.IsNull() && !state.DsDigestType.IsUnknown() {
			params.Set("digestType", fmt.Sprintf("%d", state.DsDigestType.ValueInt64()))
		}
		if !state.DsDigest.IsNull() && !state.DsDigest.IsUnknown() {
			params.Set("digest", state.DsDigest.ValueString())
		}
	case "SVCB", "HTTPS":
		if !state.SvcPriority.IsNull() && !state.SvcPriority.IsUnknown() {
			params.Set("svcPriority", fmt.Sprintf("%d", state.SvcPriority.ValueInt64()))
		}
		if !state.SvcTargetName.IsNull() && !state.SvcTargetName.IsUnknown() {
			params.Set("svcTargetName", state.SvcTargetName.ValueString())
		}
		if !state.SvcParams.IsNull() && !state.SvcParams.IsUnknown() {
			params.Set("svcParams", state.SvcParams.ValueString())
		}
	}

	return params
}

// setValueParams sets the type-specific value parameters on an Add request.
func setValueParams(params url.Values, plan *dnsRecordResourceModel) {
	recType := plan.Type.ValueString()
	value := plan.Value.ValueString()

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
		if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
			params.Set("preference", fmt.Sprintf("%d", plan.Priority.ValueInt64()))
		}
	case "TXT":
		params.Set("text", value)
	case "SRV":
		params.Set("target", value)
		if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
			params.Set("priority", fmt.Sprintf("%d", plan.Priority.ValueInt64()))
		}
		if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
			params.Set("weight", fmt.Sprintf("%d", plan.Weight.ValueInt64()))
		}
		if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
			params.Set("port", fmt.Sprintf("%d", plan.Port.ValueInt64()))
		}
	case "CAA":
		params.Set("value", value)
		if !plan.Flags.IsNull() && !plan.Flags.IsUnknown() {
			params.Set("flags", fmt.Sprintf("%d", plan.Flags.ValueInt64()))
		}
		if !plan.Tag.IsNull() && !plan.Tag.IsUnknown() {
			params.Set("tag", plan.Tag.ValueString())
		}
	case "SOA":
		params.Set("primaryNameServer", value)
	case "FWD":
		params.Set("forwarder", value)
		if !plan.Protocol.IsNull() && !plan.Protocol.IsUnknown() {
			params.Set("protocol", plan.Protocol.ValueString())
		}
		if !plan.ForwarderPriority.IsNull() && !plan.ForwarderPriority.IsUnknown() {
			params.Set("forwarderPriority", fmt.Sprintf("%d", plan.ForwarderPriority.ValueInt64()))
		}
		if !plan.DnssecValidation.IsNull() && !plan.DnssecValidation.IsUnknown() {
			params.Set("dnssecValidation", fmt.Sprintf("%t", plan.DnssecValidation.ValueBool()))
		}
		if !plan.ProxyType.IsNull() && !plan.ProxyType.IsUnknown() {
			params.Set("proxyType", plan.ProxyType.ValueString())
		}
		if !plan.ProxyAddress.IsNull() && !plan.ProxyAddress.IsUnknown() {
			params.Set("proxyAddress", plan.ProxyAddress.ValueString())
		}
		if !plan.ProxyPort.IsNull() && !plan.ProxyPort.IsUnknown() {
			params.Set("proxyPort", fmt.Sprintf("%d", plan.ProxyPort.ValueInt64()))
		}
		if !plan.ProxyUsername.IsNull() && !plan.ProxyUsername.IsUnknown() {
			params.Set("proxyUsername", plan.ProxyUsername.ValueString())
		}
		if !plan.ProxyPassword.IsNull() && !plan.ProxyPassword.IsUnknown() {
			params.Set("proxyPassword", plan.ProxyPassword.ValueString())
		}
	case "APP":
		if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
			params.Set("appName", plan.AppName.ValueString())
		}
		if !plan.ClassPath.IsNull() && !plan.ClassPath.IsUnknown() {
			params.Set("classPath", plan.ClassPath.ValueString())
		}
		if !plan.RecordData.IsNull() && !plan.RecordData.IsUnknown() {
			params.Set("recordData", plan.RecordData.ValueString())
		}
	case "ANAME":
		params.Set("aname", value)
	case "DNAME":
		params.Set("dname", value)
	case "NAPTR":
		if !plan.NaptrReplacement.IsNull() && !plan.NaptrReplacement.IsUnknown() {
			params.Set("naptrReplacement", plan.NaptrReplacement.ValueString())
		}
		if !plan.NaptrOrder.IsNull() && !plan.NaptrOrder.IsUnknown() {
			params.Set("naptrOrder", fmt.Sprintf("%d", plan.NaptrOrder.ValueInt64()))
		}
		if !plan.NaptrPreference.IsNull() && !plan.NaptrPreference.IsUnknown() {
			params.Set("naptrPreference", fmt.Sprintf("%d", plan.NaptrPreference.ValueInt64()))
		}
		if !plan.NaptrFlags.IsNull() && !plan.NaptrFlags.IsUnknown() {
			params.Set("naptrFlags", plan.NaptrFlags.ValueString())
		}
		if !plan.NaptrServices.IsNull() && !plan.NaptrServices.IsUnknown() {
			params.Set("naptrServices", plan.NaptrServices.ValueString())
		}
		if !plan.NaptrRegexp.IsNull() && !plan.NaptrRegexp.IsUnknown() {
			params.Set("naptrRegexp", plan.NaptrRegexp.ValueString())
		}
	case "SSHFP":
		if !plan.SshfpFingerprint.IsNull() && !plan.SshfpFingerprint.IsUnknown() {
			params.Set("sshfpFingerprint", plan.SshfpFingerprint.ValueString())
		}
		if !plan.SshfpAlgorithm.IsNull() && !plan.SshfpAlgorithm.IsUnknown() {
			params.Set("sshfpAlgorithm", fmt.Sprintf("%d", plan.SshfpAlgorithm.ValueInt64()))
		}
		if !plan.SshfpFingerprintType.IsNull() && !plan.SshfpFingerprintType.IsUnknown() {
			params.Set("sshfpFingerprintType", fmt.Sprintf("%d", plan.SshfpFingerprintType.ValueInt64()))
		}
	case "TLSA":
		if !plan.TlsaCertificateAssociationData.IsNull() && !plan.TlsaCertificateAssociationData.IsUnknown() {
			params.Set("tlsaCertificateAssociationData", plan.TlsaCertificateAssociationData.ValueString())
		}
		if !plan.TlsaCertificateUsage.IsNull() && !plan.TlsaCertificateUsage.IsUnknown() {
			params.Set("tlsaCertificateUsage", plan.TlsaCertificateUsage.ValueString())
		}
		if !plan.TlsaSelector.IsNull() && !plan.TlsaSelector.IsUnknown() {
			params.Set("tlsaSelector", plan.TlsaSelector.ValueString())
		}
		if !plan.TlsaMatchingType.IsNull() && !plan.TlsaMatchingType.IsUnknown() {
			params.Set("tlsaMatchingType", plan.TlsaMatchingType.ValueString())
		}
	case "URI":
		if !plan.Uri.IsNull() && !plan.Uri.IsUnknown() {
			params.Set("uri", plan.Uri.ValueString())
		}
		if !plan.UriPriority.IsNull() && !plan.UriPriority.IsUnknown() {
			params.Set("uriPriority", fmt.Sprintf("%d", plan.UriPriority.ValueInt64()))
		}
		if !plan.UriWeight.IsNull() && !plan.UriWeight.IsUnknown() {
			params.Set("uriWeight", fmt.Sprintf("%d", plan.UriWeight.ValueInt64()))
		}
	case "DS":
		if !plan.DsDigest.IsNull() && !plan.DsDigest.IsUnknown() {
			params.Set("digest", plan.DsDigest.ValueString())
		}
		if !plan.DsKeyTag.IsNull() && !plan.DsKeyTag.IsUnknown() {
			params.Set("keyTag", fmt.Sprintf("%d", plan.DsKeyTag.ValueInt64()))
		}
		if !plan.DsAlgorithm.IsNull() && !plan.DsAlgorithm.IsUnknown() {
			params.Set("algorithm", fmt.Sprintf("%d", plan.DsAlgorithm.ValueInt64()))
		}
		if !plan.DsDigestType.IsNull() && !plan.DsDigestType.IsUnknown() {
			params.Set("digestType", fmt.Sprintf("%d", plan.DsDigestType.ValueInt64()))
		}
	case "SVCB", "HTTPS":
		if !plan.SvcTargetName.IsNull() && !plan.SvcTargetName.IsUnknown() {
			params.Set("svcTargetName", plan.SvcTargetName.ValueString())
		}
		if !plan.SvcPriority.IsNull() && !plan.SvcPriority.IsUnknown() {
			params.Set("svcPriority", fmt.Sprintf("%d", plan.SvcPriority.ValueInt64()))
		}
		if !plan.SvcParams.IsNull() && !plan.SvcParams.IsUnknown() {
			params.Set("svcParams", plan.SvcParams.ValueString())
		}
		if !plan.AutoIpv4Hint.IsNull() && !plan.AutoIpv4Hint.IsUnknown() {
			params.Set("autoIpv4Hint", fmt.Sprintf("%t", plan.AutoIpv4Hint.ValueBool()))
		}
		if !plan.AutoIpv6Hint.IsNull() && !plan.AutoIpv6Hint.IsUnknown() {
			params.Set("autoIpv6Hint", fmt.Sprintf("%t", plan.AutoIpv6Hint.ValueBool()))
		}
	}
}

// compositeID builds the composite identifier from its components.
func compositeID(zone, domain, recordType, value string) string {
	return fmt.Sprintf("%s:%s:%s:%s", zone, domain, recordType, value)
}

func sshfpAlgorithmToInt(name string) int64 {
	switch name {
	case "RSA":
		return 1
	case "DSA":
		return 2
	case "ECDSA":
		return 3
	case "Ed25519":
		return 4
	default:
		return 0
	}
}

func sshfpFingerprintTypeToInt(name string) int64 {
	switch name {
	case "SHA1":
		return 1
	case "SHA256":
		return 2
	default:
		return 0
	}
}
