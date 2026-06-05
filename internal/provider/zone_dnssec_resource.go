package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var (
	_ resource.Resource                = &zoneDnssecResource{}
	_ resource.ResourceWithImportState = &zoneDnssecResource{}
)

type zoneDnssecResource struct {
	client *client.Client
}

type zoneDnssecResourceModel struct {
	Zone            types.String `tfsdk:"zone"`
	Algorithm       types.String `tfsdk:"algorithm"`
	HashAlgorithm   types.String `tfsdk:"hash_algorithm"`
	KskKeySize      types.Int64  `tfsdk:"ksk_key_size"`
	ZskKeySize      types.Int64  `tfsdk:"zsk_key_size"`
	Curve           types.String `tfsdk:"curve"`
	DnsKeyTtl       types.Int64  `tfsdk:"dnskey_ttl"`
	ZskRolloverDays types.Int64  `tfsdk:"zsk_rollover_days"`
	NxProof         types.String `tfsdk:"nx_proof"`
	Iterations      types.Int64  `tfsdk:"iterations"`
	SaltLength      types.Int64  `tfsdk:"salt_length"`
	DnssecStatus    types.String `tfsdk:"dnssec_status"`
}

func NewZoneDnssecResource() resource.Resource {
	return &zoneDnssecResource{}
}

func (r *zoneDnssecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone_dnssec"
}

func (r *zoneDnssecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DNSSEC signing for an authoritative zone on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Description: "The zone name to sign with DNSSEC.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Description: "DNSSEC algorithm: RSA, ECDSA, or EDDSA.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("RSA", "ECDSA", "EDDSA"),
				},
			},
			"hash_algorithm": schema.StringAttribute{
				Description: "Hash algorithm for RSA: MD5, SHA1, SHA256, SHA512.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("MD5", "SHA1", "SHA256", "SHA512"),
				},
			},
			"ksk_key_size": schema.Int64Attribute{
				Description: "Key Signing Key size in bits (for RSA).",
				Optional:    true,
			},
			"zsk_key_size": schema.Int64Attribute{
				Description: "Zone Signing Key size in bits (for RSA).",
				Optional:    true,
			},
			"curve": schema.StringAttribute{
				Description: "Elliptic curve: P256, P384 (ECDSA) or ED25519, ED448 (EDDSA).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("P256", "P384", "ED25519", "ED448"),
				},
			},
			"dnskey_ttl": schema.Int64Attribute{
				Description: "TTL for DNSKEY records in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(86400),
			},
			"zsk_rollover_days": schema.Int64Attribute{
				Description: "Number of days between ZSK rollovers.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
			},
			"nx_proof": schema.StringAttribute{
				Description: "Non-existence proof type: NSEC or NSEC3.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("NSEC", "NSEC3"),
				},
			},
			"iterations": schema.Int64Attribute{
				Description: "NSEC3 iterations.",
				Optional:    true,
			},
			"salt_length": schema.Int64Attribute{
				Description: "NSEC3 salt length.",
				Optional:    true,
			},
			"dnssec_status": schema.StringAttribute{
				Description: "Current DNSSEC status of the zone (read from zone options).",
				Computed:    true,
			},
		},
	}
}

func (r *zoneDnssecResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *zoneDnssecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan zoneDnssecResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := plan.Zone.ValueString()

	tflog.Debug(ctx, "Signing zone with DNSSEC", map[string]interface{}{
		"zone": zone,
	})

	params := r.buildSignParams(&plan)

	_, err := r.client.SignZone(zone, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Signing Zone",
			fmt.Sprintf("Could not sign zone %q: %s", zone, err),
		)
		return
	}

	// Read back DNSSEC properties.
	r.readDnssecIntoModel(ctx, zone, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneDnssecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state zoneDnssecResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()

	// Check zone options to see if DNSSEC is still active.
	zoneOpts, err := r.client.GetZoneOptions(zone)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	dnssecStatus, _ := zoneOpts["dnssecStatus"].(string)
	if dnssecStatus == "" || dnssecStatus == "Unsigned" {
		resp.State.RemoveResource(ctx)
		return
	}

	r.readDnssecIntoModel(ctx, zone, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *zoneDnssecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan zoneDnssecResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := plan.Zone.ValueString()

	tflog.Debug(ctx, "Updating zone DNSSEC (unsign then re-sign)", map[string]interface{}{
		"zone": zone,
	})

	// DNSSEC params cannot be updated in place; unsign then re-sign.
	_, err := r.client.UnsignZone(zone)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Unsigning Zone",
			fmt.Sprintf("Could not unsign zone %q for re-signing: %s", zone, err),
		)
		return
	}

	params := r.buildSignParams(&plan)

	_, err = r.client.SignZone(zone, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Re-signing Zone",
			fmt.Sprintf("Could not re-sign zone %q: %s", zone, err),
		)
		return
	}

	// Read back.
	r.readDnssecIntoModel(ctx, zone, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *zoneDnssecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state zoneDnssecResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()

	tflog.Debug(ctx, "Unsigning zone DNSSEC", map[string]interface{}{
		"zone": zone,
	})

	_, err := r.client.UnsignZone(zone)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Unsigning Zone",
			fmt.Sprintf("Could not unsign zone %q: %s", zone, err),
		)
	}
}

func (r *zoneDnssecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zone := req.ID

	model := zoneDnssecResourceModel{
		Zone: types.StringValue(zone),
	}

	r.readDnssecIntoModel(ctx, zone, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// buildSignParams constructs the url.Values for the sign API call.
func (r *zoneDnssecResource) buildSignParams(model *zoneDnssecResourceModel) url.Values {
	params := url.Values{}
	params.Set("algorithm", model.Algorithm.ValueString())

	if !model.HashAlgorithm.IsNull() && !model.HashAlgorithm.IsUnknown() {
		params.Set("hashName", model.HashAlgorithm.ValueString())
	}

	if !model.KskKeySize.IsNull() && !model.KskKeySize.IsUnknown() {
		params.Set("kskKeySize", fmt.Sprintf("%d", model.KskKeySize.ValueInt64()))
	}

	if !model.ZskKeySize.IsNull() && !model.ZskKeySize.IsUnknown() {
		params.Set("zskKeySize", fmt.Sprintf("%d", model.ZskKeySize.ValueInt64()))
	}

	if !model.Curve.IsNull() && !model.Curve.IsUnknown() {
		params.Set("curve", model.Curve.ValueString())
	}

	if !model.DnsKeyTtl.IsNull() && !model.DnsKeyTtl.IsUnknown() {
		params.Set("dnsKeyTtl", fmt.Sprintf("%d", model.DnsKeyTtl.ValueInt64()))
	}

	if !model.ZskRolloverDays.IsNull() && !model.ZskRolloverDays.IsUnknown() {
		params.Set("zskRolloverDays", fmt.Sprintf("%d", model.ZskRolloverDays.ValueInt64()))
	}

	if !model.NxProof.IsNull() && !model.NxProof.IsUnknown() {
		params.Set("nxProof", model.NxProof.ValueString())
	}

	if !model.Iterations.IsNull() && !model.Iterations.IsUnknown() {
		params.Set("iterations", fmt.Sprintf("%d", model.Iterations.ValueInt64()))
	}

	if !model.SaltLength.IsNull() && !model.SaltLength.IsUnknown() {
		params.Set("saltLength", fmt.Sprintf("%d", model.SaltLength.ValueInt64()))
	}

	return params
}

// readDnssecIntoModel reads DNSSEC properties and zone options into the model.
func (r *zoneDnssecResource) readDnssecIntoModel(_ context.Context, zone string, model *zoneDnssecResourceModel, diags *diag.Diagnostics) {
	// Get DNSSEC properties.
	props, err := r.client.GetDNSSECProperties(zone)
	if err != nil {
		diags.AddError(
			"Error Reading DNSSEC Properties",
			fmt.Sprintf("Could not read DNSSEC properties for zone %q: %s", zone, err),
		)
		return
	}

	if algo, ok := props["algorithm"].(string); ok {
		model.Algorithm = types.StringValue(algo)
	}

	if hashAlgo, ok := props["hashName"].(string); ok && hashAlgo != "" {
		model.HashAlgorithm = types.StringValue(hashAlgo)
	}

	if ksk, ok := props["kskKeySize"].(float64); ok {
		model.KskKeySize = types.Int64Value(int64(ksk))
	}

	if zsk, ok := props["zskKeySize"].(float64); ok {
		model.ZskKeySize = types.Int64Value(int64(zsk))
	}

	if curve, ok := props["curve"].(string); ok && curve != "" {
		model.Curve = types.StringValue(curve)
	}

	if ttl, ok := props["dnsKeyTtl"].(float64); ok {
		model.DnsKeyTtl = types.Int64Value(int64(ttl))
	}

	if rollover, ok := props["zskRolloverDays"].(float64); ok {
		model.ZskRolloverDays = types.Int64Value(int64(rollover))
	}

	if nxProof, ok := props["nxProof"].(string); ok && nxProof != "" {
		model.NxProof = types.StringValue(nxProof)
	}

	if iter, ok := props["iterations"].(float64); ok {
		model.Iterations = types.Int64Value(int64(iter))
	}

	if salt, ok := props["saltLength"].(float64); ok {
		model.SaltLength = types.Int64Value(int64(salt))
	}

	// Get zone options for dnssecStatus.
	zoneOpts, err := r.client.GetZoneOptions(zone)
	if err != nil {
		diags.AddError(
			"Error Reading Zone Options",
			fmt.Sprintf("Could not read zone options for %q: %s", zone, err),
		)
		return
	}

	if status, ok := zoneOpts["dnssecStatus"].(string); ok {
		model.DnssecStatus = types.StringValue(status)
	}
}
