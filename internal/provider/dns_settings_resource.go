package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ resource.Resource = &dnsSettingsResource{}

func NewDNSSettingsResource() resource.Resource {
	return &dnsSettingsResource{}
}

type dnsSettingsResource struct {
	client *client.Client
}

type dnsSettingsResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	DnsServerDomain       types.String `tfsdk:"dns_server_domain"`
	DefaultRecordTtl      types.Int64  `tfsdk:"default_record_ttl"`
	PreferIPv6            types.Bool   `tfsdk:"prefer_ipv6"`
	DnssecValidation      types.Bool   `tfsdk:"dnssec_validation"`
	QnameMinimization     types.Bool   `tfsdk:"qname_minimization"`
	RandomizeName         types.Bool   `tfsdk:"randomize_name"`
	Recursion             types.String `tfsdk:"recursion"`
	ServeStale            types.Bool   `tfsdk:"serve_stale"`
	CacheMaximumEntries   types.Int64  `tfsdk:"cache_maximum_entries"`
	CacheMinimumRecordTtl types.Int64  `tfsdk:"cache_minimum_record_ttl"`
	CacheMaximumRecordTtl types.Int64  `tfsdk:"cache_maximum_record_ttl"`
	CacheNegativeRecordTtl types.Int64 `tfsdk:"cache_negative_record_ttl"`
	EnableBlocking        types.Bool   `tfsdk:"enable_blocking"`
	BlockingType          types.String `tfsdk:"blocking_type"`
	BlockListUrls         types.List   `tfsdk:"block_list_urls"`
	Forwarders            types.List   `tfsdk:"forwarders"`
	ForwarderProtocol     types.String `tfsdk:"forwarder_protocol"`
	EnableLogging         types.Bool   `tfsdk:"enable_logging"`
	LogQueries            types.Bool   `tfsdk:"log_queries"`
	MaxLogFileDays        types.Int64  `tfsdk:"max_log_file_days"`
}

func (r *dnsSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
}

func (r *dnsSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages DNS server settings on a Technitium DNS Server. This is a singleton resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for the settings resource. Always 'settings'.",
				Computed:    true,
			},
			"dns_server_domain": schema.StringAttribute{
				Description: "The primary domain name used by the DNS server.",
				Optional:    true,
				Computed:    true,
			},
			"default_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for new records.",
				Optional:    true,
				Computed:    true,
			},
			"prefer_ipv6": schema.BoolAttribute{
				Description: "Prefer IPv6 for DNS resolution.",
				Optional:    true,
				Computed:    true,
			},
			"dnssec_validation": schema.BoolAttribute{
				Description: "Enable DNSSEC validation for DNS responses.",
				Optional:    true,
				Computed:    true,
			},
			"qname_minimization": schema.BoolAttribute{
				Description: "Enable QNAME minimization for recursive queries.",
				Optional:    true,
				Computed:    true,
			},
			"randomize_name": schema.BoolAttribute{
				Description: "Randomize query name casing for cache poisoning protection.",
				Optional:    true,
				Computed:    true,
			},
			"recursion": schema.StringAttribute{
				Description: "Recursion policy. Valid values: Allow, Deny, AllowOnlyForPrivateNetworks, UseSpecifiedNetworkACL.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Allow", "Deny", "AllowOnlyForPrivateNetworks", "UseSpecifiedNetworkACL"),
				},
			},
			"serve_stale": schema.BoolAttribute{
				Description: "Serve stale cached records when upstream is unavailable.",
				Optional:    true,
				Computed:    true,
			},
			"cache_maximum_entries": schema.Int64Attribute{
				Description: "Maximum number of entries in the DNS cache.",
				Optional:    true,
				Computed:    true,
			},
			"cache_minimum_record_ttl": schema.Int64Attribute{
				Description: "Minimum TTL in seconds for cached records.",
				Optional:    true,
				Computed:    true,
			},
			"cache_maximum_record_ttl": schema.Int64Attribute{
				Description: "Maximum TTL in seconds for cached records.",
				Optional:    true,
				Computed:    true,
			},
			"cache_negative_record_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for negative (NXDOMAIN) cache entries.",
				Optional:    true,
				Computed:    true,
			},
			"enable_blocking": schema.BoolAttribute{
				Description: "Enable DNS-level ad/malware blocking.",
				Optional:    true,
				Computed:    true,
			},
			"blocking_type": schema.StringAttribute{
				Description: "How blocked queries are answered. Valid values: AnyAddress, NxDomain, CustomAddress.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("AnyAddress", "NxDomain", "CustomAddress"),
				},
			},
			"block_list_urls": schema.ListAttribute{
				Description: "URLs of DNS block lists.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"forwarders": schema.ListAttribute{
				Description: "List of forwarder DNS server addresses.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"forwarder_protocol": schema.StringAttribute{
				Description: "Protocol for DNS forwarding. Valid values: Udp, Tcp, Tls, Https, Quic.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Udp", "Tcp", "Tls", "Https", "Quic"),
				},
			},
			"enable_logging": schema.BoolAttribute{
				Description: "Enable DNS query logging.",
				Optional:    true,
				Computed:    true,
			},
			"log_queries": schema.BoolAttribute{
				Description: "Log all DNS queries.",
				Optional:    true,
				Computed:    true,
			},
			"max_log_file_days": schema.Int64Attribute{
				Description: "Number of days to retain log files.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *dnsSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dnsSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildSetParams(ctx, &plan)

	tflog.Debug(ctx, "Applying DNS settings")

	if len(params) > 0 {
		_, err := r.client.SetDNSSettings(params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting DNS Settings",
				fmt.Sprintf("Could not apply DNS settings: %s", err),
			)
			return
		}
	}

	plan.ID = types.StringValue("settings")

	diags := r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readIntoModel(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := r.buildSetParams(ctx, &plan)

	tflog.Debug(ctx, "Updating DNS settings")

	if len(params) > 0 {
		_, err := r.client.SetDNSSettings(params)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating DNS Settings",
				fmt.Sprintf("Could not update DNS settings: %s", err),
			)
			return
		}
	}

	plan.ID = types.StringValue("settings")

	diags := r.readIntoModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// DNS settings cannot be deleted. Removing from state only.
}

func (r *dnsSettingsResource) readIntoModel(ctx context.Context, model *dnsSettingsResourceModel) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "Reading DNS settings")

	response, err := r.client.GetDNSSettings()
	if err != nil {
		diags.AddError("Error Reading DNS Settings", fmt.Sprintf("Could not read DNS settings: %s", err))
		return
	}

	model.ID = types.StringValue("settings")

	if v, ok := response["dnsServerDomain"].(string); ok {
		model.DnsServerDomain = types.StringValue(v)
	}
	if v, ok := response["defaultRecordTtl"].(float64); ok {
		model.DefaultRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["preferIPv6"].(bool); ok {
		model.PreferIPv6 = types.BoolValue(v)
	}
	if v, ok := response["dnssecValidation"].(bool); ok {
		model.DnssecValidation = types.BoolValue(v)
	}
	if v, ok := response["qnameMinimization"].(bool); ok {
		model.QnameMinimization = types.BoolValue(v)
	}
	if v, ok := response["randomizeName"].(bool); ok {
		model.RandomizeName = types.BoolValue(v)
	}
	if v, ok := response["recursion"].(string); ok {
		model.Recursion = types.StringValue(v)
	}
	if v, ok := response["serveStale"].(bool); ok {
		model.ServeStale = types.BoolValue(v)
	}
	if v, ok := response["cacheMaximumEntries"].(float64); ok {
		model.CacheMaximumEntries = types.Int64Value(int64(v))
	}
	if v, ok := response["cacheMinimumRecordTtl"].(float64); ok {
		model.CacheMinimumRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["cacheMaximumRecordTtl"].(float64); ok {
		model.CacheMaximumRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["cacheNegativeRecordTtl"].(float64); ok {
		model.CacheNegativeRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["enableBlocking"].(bool); ok {
		model.EnableBlocking = types.BoolValue(v)
	}
	if v, ok := response["blockingType"].(string); ok {
		model.BlockingType = types.StringValue(v)
	}
	if v, ok := response["forwarderProtocol"].(string); ok {
		model.ForwarderProtocol = types.StringValue(v)
	}
	if v, ok := response["enableLogging"].(bool); ok {
		model.EnableLogging = types.BoolValue(v)
	}
	if v, ok := response["logQueries"].(bool); ok {
		model.LogQueries = types.BoolValue(v)
	}
	if v, ok := response["maxLogFileDays"].(float64); ok {
		model.MaxLogFileDays = types.Int64Value(int64(v))
	}

	if v, ok := response["blockListUrls"].([]interface{}); ok {
		urls := make([]string, len(v))
		for i, u := range v {
			urls[i], _ = u.(string)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, urls)
		diags.Append(d...)
		model.BlockListUrls = listVal
	} else {
		model.BlockListUrls = types.ListNull(types.StringType)
	}

	if v, ok := response["forwarders"].([]interface{}); ok {
		fwds := make([]string, len(v))
		for i, f := range v {
			fwds[i], _ = f.(string)
		}
		listVal, d := types.ListValueFrom(ctx, types.StringType, fwds)
		diags.Append(d...)
		model.Forwarders = listVal
	} else {
		model.Forwarders = types.ListNull(types.StringType)
	}

	return
}

func (r *dnsSettingsResource) buildSetParams(ctx context.Context, model *dnsSettingsResourceModel) url.Values {
	params := url.Values{}

	if !model.DnsServerDomain.IsNull() && !model.DnsServerDomain.IsUnknown() {
		params.Set("dnsServerDomain", model.DnsServerDomain.ValueString())
	}
	if !model.DefaultRecordTtl.IsNull() && !model.DefaultRecordTtl.IsUnknown() {
		params.Set("defaultRecordTtl", fmt.Sprintf("%d", model.DefaultRecordTtl.ValueInt64()))
	}
	if !model.PreferIPv6.IsNull() && !model.PreferIPv6.IsUnknown() {
		params.Set("preferIPv6", fmt.Sprintf("%t", model.PreferIPv6.ValueBool()))
	}
	if !model.DnssecValidation.IsNull() && !model.DnssecValidation.IsUnknown() {
		params.Set("dnssecValidation", fmt.Sprintf("%t", model.DnssecValidation.ValueBool()))
	}
	if !model.QnameMinimization.IsNull() && !model.QnameMinimization.IsUnknown() {
		params.Set("qnameMinimization", fmt.Sprintf("%t", model.QnameMinimization.ValueBool()))
	}
	if !model.RandomizeName.IsNull() && !model.RandomizeName.IsUnknown() {
		params.Set("randomizeName", fmt.Sprintf("%t", model.RandomizeName.ValueBool()))
	}
	if !model.Recursion.IsNull() && !model.Recursion.IsUnknown() {
		params.Set("recursion", model.Recursion.ValueString())
	}
	if !model.ServeStale.IsNull() && !model.ServeStale.IsUnknown() {
		params.Set("serveStale", fmt.Sprintf("%t", model.ServeStale.ValueBool()))
	}
	if !model.CacheMaximumEntries.IsNull() && !model.CacheMaximumEntries.IsUnknown() {
		params.Set("cacheMaximumEntries", fmt.Sprintf("%d", model.CacheMaximumEntries.ValueInt64()))
	}
	if !model.CacheMinimumRecordTtl.IsNull() && !model.CacheMinimumRecordTtl.IsUnknown() {
		params.Set("cacheMinimumRecordTtl", fmt.Sprintf("%d", model.CacheMinimumRecordTtl.ValueInt64()))
	}
	if !model.CacheMaximumRecordTtl.IsNull() && !model.CacheMaximumRecordTtl.IsUnknown() {
		params.Set("cacheMaximumRecordTtl", fmt.Sprintf("%d", model.CacheMaximumRecordTtl.ValueInt64()))
	}
	if !model.CacheNegativeRecordTtl.IsNull() && !model.CacheNegativeRecordTtl.IsUnknown() {
		params.Set("cacheNegativeRecordTtl", fmt.Sprintf("%d", model.CacheNegativeRecordTtl.ValueInt64()))
	}
	if !model.EnableBlocking.IsNull() && !model.EnableBlocking.IsUnknown() {
		params.Set("enableBlocking", fmt.Sprintf("%t", model.EnableBlocking.ValueBool()))
	}
	if !model.BlockingType.IsNull() && !model.BlockingType.IsUnknown() {
		params.Set("blockingType", model.BlockingType.ValueString())
	}
	if !model.ForwarderProtocol.IsNull() && !model.ForwarderProtocol.IsUnknown() {
		params.Set("forwarderProtocol", model.ForwarderProtocol.ValueString())
	}
	if !model.EnableLogging.IsNull() && !model.EnableLogging.IsUnknown() {
		params.Set("enableLogging", fmt.Sprintf("%t", model.EnableLogging.ValueBool()))
	}
	if !model.LogQueries.IsNull() && !model.LogQueries.IsUnknown() {
		params.Set("logQueries", fmt.Sprintf("%t", model.LogQueries.ValueBool()))
	}
	if !model.MaxLogFileDays.IsNull() && !model.MaxLogFileDays.IsUnknown() {
		params.Set("maxLogFileDays", fmt.Sprintf("%d", model.MaxLogFileDays.ValueInt64()))
	}

	if !model.BlockListUrls.IsNull() && !model.BlockListUrls.IsUnknown() {
		var urls []string
		model.BlockListUrls.ElementsAs(ctx, &urls, false)
		params.Set("blockListUrls", strings.Join(urls, ","))
	}
	if !model.Forwarders.IsNull() && !model.Forwarders.IsUnknown() {
		var fwds []string
		model.Forwarders.ElementsAs(ctx, &fwds, false)
		params.Set("forwarders", strings.Join(fwds, ","))
	}

	return params
}
