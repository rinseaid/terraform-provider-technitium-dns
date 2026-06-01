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
	_ resource.Resource                = &tsigKeyResource{}
	_ resource.ResourceWithImportState = &tsigKeyResource{}
)

type tsigKeyResource struct {
	client *client.Client
}

type tsigKeyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	KeyName      types.String `tfsdk:"key_name"`
	Algorithm    types.String `tfsdk:"algorithm"`
	SharedSecret types.String `tfsdk:"shared_secret"`
}

func NewTSIGKeyResource() resource.Resource {
	return &tsigKeyResource{}
}

func (r *tsigKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tsig_key"
}

func (r *tsigKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a TSIG key on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for the TSIG key. Same as key_name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_name": schema.StringAttribute{
				Description: "DNS name identifying the TSIG key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"algorithm": schema.StringAttribute{
				Description: "HMAC algorithm for the TSIG key.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"hmac-md5.sig-alg.reg.int",
						"hmac-sha1",
						"hmac-sha256",
						"hmac-sha256-128",
						"hmac-sha384",
						"hmac-sha384-192",
						"hmac-sha512",
						"hmac-sha512-256",
					),
				},
			},
			"shared_secret": schema.StringAttribute{
				Description: "Base64-encoded shared secret. If omitted, the server generates one.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *tsigKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tsigKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tsigKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyName := plan.KeyName.ValueString()
	algorithm := plan.Algorithm.ValueString()
	sharedSecret := plan.SharedSecret.ValueString()

	if strings.Contains(keyName, "|") {
		resp.Diagnostics.AddError(
			"Invalid Key Name",
			"key_name must not contain the pipe character '|'.",
		)
		return
	}
	if strings.Contains(algorithm, "|") {
		resp.Diagnostics.AddError(
			"Invalid Algorithm",
			"algorithm must not contain the pipe character '|'.",
		)
		return
	}
	if strings.Contains(sharedSecret, "|") {
		resp.Diagnostics.AddError(
			"Invalid Shared Secret",
			"shared_secret must not contain the pipe character '|'.",
		)
		return
	}

	tflog.Debug(ctx, "Creating TSIG key", map[string]interface{}{
		"key_name": keyName,
	})

	existingKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read existing TSIG keys: %s", err),
		)
		return
	}

	for _, k := range existingKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			resp.Diagnostics.AddError(
				"Duplicate TSIG Key",
				fmt.Sprintf("A TSIG key with name %q already exists.", keyName),
			)
			return
		}
	}

	newKey := tsigKeyEntry{
		keyName:      keyName,
		sharedSecret: sharedSecret,
		algorithm:    algorithm,
	}
	existingKeys = append(existingKeys, newKey)

	if err := r.writeAllTSIGKeys(existingKeys); err != nil {
		resp.Diagnostics.AddError(
			"Error Creating TSIG Key",
			fmt.Sprintf("Could not write TSIG keys: %s", err),
		)
		return
	}

	// Read back to pick up server-generated secret.
	allKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read TSIG keys after create: %s", err),
		)
		return
	}

	found := false
	for _, k := range allKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			plan.ID = types.StringValue(keyName)
			plan.Algorithm = types.StringValue(k.algorithm)
			plan.SharedSecret = types.StringValue(k.sharedSecret)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Key",
			fmt.Sprintf("TSIG key %q not found after creation.", keyName),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tsigKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tsigKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyName := state.KeyName.ValueString()

	allKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read TSIG keys: %s", err),
		)
		return
	}

	found := false
	for _, k := range allKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			state.ID = types.StringValue(keyName)
			state.Algorithm = types.StringValue(k.algorithm)
			state.SharedSecret = types.StringValue(k.sharedSecret)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tsigKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tsigKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyName := plan.KeyName.ValueString()
	algorithm := plan.Algorithm.ValueString()
	sharedSecret := plan.SharedSecret.ValueString()

	if strings.Contains(algorithm, "|") {
		resp.Diagnostics.AddError(
			"Invalid Algorithm",
			"algorithm must not contain the pipe character '|'.",
		)
		return
	}
	if strings.Contains(sharedSecret, "|") {
		resp.Diagnostics.AddError(
			"Invalid Shared Secret",
			"shared_secret must not contain the pipe character '|'.",
		)
		return
	}

	tflog.Debug(ctx, "Updating TSIG key", map[string]interface{}{
		"key_name": keyName,
	})

	existingKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read existing TSIG keys: %s", err),
		)
		return
	}

	found := false
	for i, k := range existingKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			existingKeys[i].algorithm = algorithm
			existingKeys[i].sharedSecret = sharedSecret
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"TSIG Key Not Found",
			fmt.Sprintf("TSIG key %q does not exist on the server.", keyName),
		)
		return
	}

	if err := r.writeAllTSIGKeys(existingKeys); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating TSIG Key",
			fmt.Sprintf("Could not write TSIG keys: %s", err),
		)
		return
	}

	// Read back to confirm.
	allKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read TSIG keys after update: %s", err),
		)
		return
	}

	for _, k := range allKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			plan.ID = types.StringValue(keyName)
			plan.Algorithm = types.StringValue(k.algorithm)
			plan.SharedSecret = types.StringValue(k.sharedSecret)
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tsigKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tsigKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyName := state.KeyName.ValueString()

	tflog.Debug(ctx, "Deleting TSIG key", map[string]interface{}{
		"key_name": keyName,
	})

	existingKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading TSIG Keys",
			fmt.Sprintf("Could not read existing TSIG keys: %s", err),
		)
		return
	}

	filtered := make([]tsigKeyEntry, 0, len(existingKeys))
	for _, k := range existingKeys {
		if normalizeTSIGKeyName(k.keyName) != normalizeTSIGKeyName(keyName) {
			filtered = append(filtered, k)
		}
	}

	if err := r.writeAllTSIGKeys(filtered); err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting TSIG Key",
			fmt.Sprintf("Could not write TSIG keys after removal: %s", err),
		)
	}
}

func (r *tsigKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	keyName := req.ID

	allKeys, err := r.readAllTSIGKeys()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Importing TSIG Key",
			fmt.Sprintf("Could not read TSIG keys: %s", err),
		)
		return
	}

	for _, k := range allKeys {
		if normalizeTSIGKeyName(k.keyName) == normalizeTSIGKeyName(keyName) {
			model := tsigKeyResourceModel{
				ID:           types.StringValue(keyName),
				KeyName:      types.StringValue(keyName),
				Algorithm:    types.StringValue(k.algorithm),
				SharedSecret: types.StringValue(k.sharedSecret),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Error Importing TSIG Key",
		fmt.Sprintf("TSIG key %q not found on the server.", keyName),
	)
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

type tsigKeyEntry struct {
	keyName      string
	sharedSecret string
	algorithm    string
}

func normalizeTSIGKeyName(name string) string {
	return strings.TrimRight(strings.ToLower(name), ".")
}

// readAllTSIGKeys fetches DNS settings and parses the tsigKeys array.
func (r *tsigKeyResource) readAllTSIGKeys() ([]tsigKeyEntry, error) {
	response, err := r.client.GetDNSSettings()
	if err != nil {
		return nil, fmt.Errorf("GetDNSSettings: %w", err)
	}

	rawKeys, ok := response["tsigKeys"].([]interface{})
	if !ok || rawKeys == nil {
		return nil, nil
	}

	keys := make([]tsigKeyEntry, 0, len(rawKeys))
	for _, raw := range rawKeys {
		entry, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		kn, _ := entry["keyName"].(string)
		ss, _ := entry["sharedSecret"].(string)
		an, _ := entry["algorithmName"].(string)
		keys = append(keys, tsigKeyEntry{
			keyName:      kn,
			sharedSecret: ss,
			algorithm:    an,
		})
	}

	return keys, nil
}

// writeAllTSIGKeys encodes the key list into the pipe-delimited wire format
// and writes it back via SetDNSSettings. An empty list sets tsigKeys=false
// to remove all keys.
func (r *tsigKeyResource) writeAllTSIGKeys(keys []tsigKeyEntry) error {
	params := url.Values{}

	if len(keys) == 0 {
		params.Set("tsigKeys", "false")
	} else {
		parts := make([]string, 0, len(keys)*3)
		for _, k := range keys {
			parts = append(parts, k.keyName, k.sharedSecret, k.algorithm)
		}
		params.Set("tsigKeys", strings.Join(parts, "|"))
	}

	_, err := r.client.SetDNSSettings(params)
	if err != nil {
		return fmt.Errorf("SetDNSSettings: %w", err)
	}

	return nil
}
