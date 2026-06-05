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
	ID                     types.String `tfsdk:"id"`
	DnsServerDomain        types.String `tfsdk:"dns_server_domain"`
	DefaultRecordTtl       types.Int64  `tfsdk:"default_record_ttl"`
	PreferIPv6             types.Bool   `tfsdk:"prefer_ipv6"`
	DnssecValidation       types.Bool   `tfsdk:"dnssec_validation"`
	QnameMinimization      types.Bool   `tfsdk:"qname_minimization"`
	RandomizeName          types.Bool   `tfsdk:"randomize_name"`
	Recursion              types.String `tfsdk:"recursion"`
	ServeStale             types.Bool   `tfsdk:"serve_stale"`
	CacheMaximumEntries    types.Int64  `tfsdk:"cache_maximum_entries"`
	CacheMinimumRecordTtl  types.Int64  `tfsdk:"cache_minimum_record_ttl"`
	CacheMaximumRecordTtl  types.Int64  `tfsdk:"cache_maximum_record_ttl"`
	CacheNegativeRecordTtl types.Int64  `tfsdk:"cache_negative_record_ttl"`
	EnableBlocking         types.Bool   `tfsdk:"enable_blocking"`
	BlockingType           types.String `tfsdk:"blocking_type"`
	BlockListUrls          types.List   `tfsdk:"block_list_urls"`
	Forwarders             types.List   `tfsdk:"forwarders"`
	ForwarderProtocol      types.String `tfsdk:"forwarder_protocol"`
	EnableLogging          types.Bool   `tfsdk:"enable_logging"`
	LogQueries             types.Bool   `tfsdk:"log_queries"`
	MaxLogFileDays         types.Int64  `tfsdk:"max_log_file_days"`
	AllowTxtBlockingReport                   types.Bool   `tfsdk:"allow_txt_blocking_report"`
	BlockingAnswerTtl                        types.Int64  `tfsdk:"blocking_answer_ttl"`
	BlockListUpdateIntervalHours             types.Int64  `tfsdk:"block_list_update_interval_hours"`
	CachePrefetchEligibility                 types.Int64  `tfsdk:"cache_prefetch_eligibility"`
	CachePrefetchTrigger                     types.Int64  `tfsdk:"cache_prefetch_trigger"`
	CachePrefetchSampleIntervalMinutes       types.Int64  `tfsdk:"cache_prefetch_sample_interval_minutes"`
	CachePrefetchSampleEligibilityHitsPerHour types.Int64 `tfsdk:"cache_prefetch_sample_eligibility_hits_per_hour"`
	CacheFailureRecordTtl                    types.Int64  `tfsdk:"cache_failure_record_ttl"`
	SaveCache                                types.Bool   `tfsdk:"save_cache"`
	ServeStaleTtl                            types.Int64  `tfsdk:"serve_stale_ttl"`
	ServeStaleAnswerTtl                      types.Int64  `tfsdk:"serve_stale_answer_ttl"`
	ServeStaleMaxWaitTime                    types.Int64  `tfsdk:"serve_stale_max_wait_time"`
	ServeStaleResetTtl                       types.Int64  `tfsdk:"serve_stale_reset_ttl"`
	ForwarderRetries                         types.Int64  `tfsdk:"forwarder_retries"`
	ForwarderTimeout                         types.Int64  `tfsdk:"forwarder_timeout"`
	ForwarderConcurrency                     types.Int64  `tfsdk:"forwarder_concurrency"`
	ConcurrentForwarding                     types.Bool   `tfsdk:"concurrent_forwarding"`
	ResolverRetries                          types.Int64  `tfsdk:"resolver_retries"`
	ResolverTimeout                          types.Int64  `tfsdk:"resolver_timeout"`
	ResolverConcurrency                      types.Int64  `tfsdk:"resolver_concurrency"`
	ResolverMaxStackCount                    types.Int64  `tfsdk:"resolver_max_stack_count"`
	ClientTimeout                            types.Int64  `tfsdk:"client_timeout"`
	UdpPayloadSize                           types.Int64  `tfsdk:"udp_payload_size"`
	TcpReceiveTimeout                        types.Int64  `tfsdk:"tcp_receive_timeout"`
	TcpSendTimeout                           types.Int64  `tfsdk:"tcp_send_timeout"`
	UdpReceiveBufferSizeKb                   types.Int64  `tfsdk:"udp_receive_buffer_size_kb"`
	UdpSendBufferSizeKb                      types.Int64  `tfsdk:"udp_send_buffer_size_kb"`
	Ipv6Mode                                 types.String `tfsdk:"ipv6_mode"`
	ListenBacklog                            types.Int64  `tfsdk:"listen_backlog"`
	DefaultSoaRecordTtl                      types.Int64  `tfsdk:"default_soa_record_ttl"`
	DefaultNsRecordTtl                       types.Int64  `tfsdk:"default_ns_record_ttl"`
	MinSoaRefresh                            types.Int64  `tfsdk:"min_soa_refresh"`
	MinSoaRetry                              types.Int64  `tfsdk:"min_soa_retry"`
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
			"allow_txt_blocking_report": schema.BoolAttribute{
				Description: "Allow TXT record blocking report queries.",
				Optional:    true,
				Computed:    true,
			},
			"blocking_answer_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for blocked DNS responses.",
				Optional:    true,
				Computed:    true,
			},
			"block_list_update_interval_hours": schema.Int64Attribute{
				Description: "Interval in hours between block list updates.",
				Optional:    true,
				Computed:    true,
			},
			"cache_prefetch_eligibility": schema.Int64Attribute{
				Description: "Minimum number of hits for a record to be eligible for cache prefetch.",
				Optional:    true,
				Computed:    true,
			},
			"cache_prefetch_trigger": schema.Int64Attribute{
				Description: "Number of hits to trigger a cache prefetch.",
				Optional:    true,
				Computed:    true,
			},
			"cache_prefetch_sample_interval_minutes": schema.Int64Attribute{
				Description: "Interval in minutes between cache prefetch sampling.",
				Optional:    true,
				Computed:    true,
			},
			"cache_prefetch_sample_eligibility_hits_per_hour": schema.Int64Attribute{
				Description: "Minimum hits per hour for a record to be eligible for prefetch sampling.",
				Optional:    true,
				Computed:    true,
			},
			"cache_failure_record_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for cached failure records.",
				Optional:    true,
				Computed:    true,
			},
			"save_cache": schema.BoolAttribute{
				Description: "Save DNS cache to disk on shutdown.",
				Optional:    true,
				Computed:    true,
			},
			"serve_stale_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for serve stale records.",
				Optional:    true,
				Computed:    true,
			},
			"serve_stale_answer_ttl": schema.Int64Attribute{
				Description: "TTL in seconds used in stale answers.",
				Optional:    true,
				Computed:    true,
			},
			"serve_stale_max_wait_time": schema.Int64Attribute{
				Description: "Maximum wait time in milliseconds before serving stale.",
				Optional:    true,
				Computed:    true,
			},
			"serve_stale_reset_ttl": schema.Int64Attribute{
				Description: "TTL in seconds to reset serve stale timer.",
				Optional:    true,
				Computed:    true,
			},
			"forwarder_retries": schema.Int64Attribute{
				Description: "Number of retries for forwarder queries.",
				Optional:    true,
				Computed:    true,
			},
			"forwarder_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for forwarder queries.",
				Optional:    true,
				Computed:    true,
			},
			"forwarder_concurrency": schema.Int64Attribute{
				Description: "Number of concurrent forwarder queries.",
				Optional:    true,
				Computed:    true,
			},
			"concurrent_forwarding": schema.BoolAttribute{
				Description: "Enable concurrent forwarding to all configured forwarders.",
				Optional:    true,
				Computed:    true,
			},
			"resolver_retries": schema.Int64Attribute{
				Description: "Number of retries for resolver queries.",
				Optional:    true,
				Computed:    true,
			},
			"resolver_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for resolver queries.",
				Optional:    true,
				Computed:    true,
			},
			"resolver_concurrency": schema.Int64Attribute{
				Description: "Number of concurrent resolver queries.",
				Optional:    true,
				Computed:    true,
			},
			"resolver_max_stack_count": schema.Int64Attribute{
				Description: "Maximum number of resolver stack entries.",
				Optional:    true,
				Computed:    true,
			},
			"client_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for client connections.",
				Optional:    true,
				Computed:    true,
			},
			"udp_payload_size": schema.Int64Attribute{
				Description: "Maximum UDP payload size in bytes.",
				Optional:    true,
				Computed:    true,
			},
			"tcp_receive_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for TCP receive operations.",
				Optional:    true,
				Computed:    true,
			},
			"tcp_send_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for TCP send operations.",
				Optional:    true,
				Computed:    true,
			},
			"udp_receive_buffer_size_kb": schema.Int64Attribute{
				Description: "UDP receive buffer size in kilobytes.",
				Optional:    true,
				Computed:    true,
			},
			"udp_send_buffer_size_kb": schema.Int64Attribute{
				Description: "UDP send buffer size in kilobytes.",
				Optional:    true,
				Computed:    true,
			},
			"ipv6_mode": schema.StringAttribute{
				Description: "IPv6 mode. Valid values: Disabled, Enabled, Preferred.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Disabled", "Enabled", "Preferred"),
				},
			},
			"listen_backlog": schema.Int64Attribute{
				Description: "TCP listen backlog size.",
				Optional:    true,
				Computed:    true,
			},
			"default_soa_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for SOA records.",
				Optional:    true,
				Computed:    true,
			},
			"default_ns_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for NS records.",
				Optional:    true,
				Computed:    true,
			},
			"min_soa_refresh": schema.Int64Attribute{
				Description: "Minimum SOA refresh interval in seconds.",
				Optional:    true,
				Computed:    true,
			},
			"min_soa_retry": schema.Int64Attribute{
				Description: "Minimum SOA retry interval in seconds.",
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

	if v, ok := response["allowTxtBlockingReport"].(bool); ok {
		model.AllowTxtBlockingReport = types.BoolValue(v)
	}
	if v, ok := response["blockingAnswerTtl"].(float64); ok {
		model.BlockingAnswerTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["blockListUpdateIntervalHours"].(float64); ok {
		model.BlockListUpdateIntervalHours = types.Int64Value(int64(v))
	}
	if v, ok := response["cachePrefetchEligibility"].(float64); ok {
		model.CachePrefetchEligibility = types.Int64Value(int64(v))
	}
	if v, ok := response["cachePrefetchTrigger"].(float64); ok {
		model.CachePrefetchTrigger = types.Int64Value(int64(v))
	}
	if v, ok := response["cachePrefetchSampleIntervalInMinutes"].(float64); ok {
		model.CachePrefetchSampleIntervalMinutes = types.Int64Value(int64(v))
	}
	if v, ok := response["cachePrefetchSampleEligibilityHitsPerHour"].(float64); ok {
		model.CachePrefetchSampleEligibilityHitsPerHour = types.Int64Value(int64(v))
	}
	if v, ok := response["cacheFailureRecordTtl"].(float64); ok {
		model.CacheFailureRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["saveCache"].(bool); ok {
		model.SaveCache = types.BoolValue(v)
	}
	if v, ok := response["serveStaleTtl"].(float64); ok {
		model.ServeStaleTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["serveStaleAnswerTtl"].(float64); ok {
		model.ServeStaleAnswerTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["serveStaleMaxWaitTime"].(float64); ok {
		model.ServeStaleMaxWaitTime = types.Int64Value(int64(v))
	}
	if v, ok := response["serveStaleResetTtl"].(float64); ok {
		model.ServeStaleResetTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["forwarderRetries"].(float64); ok {
		model.ForwarderRetries = types.Int64Value(int64(v))
	}
	if v, ok := response["forwarderTimeout"].(float64); ok {
		model.ForwarderTimeout = types.Int64Value(int64(v))
	}
	if v, ok := response["forwarderConcurrency"].(float64); ok {
		model.ForwarderConcurrency = types.Int64Value(int64(v))
	}
	if v, ok := response["concurrentForwarding"].(bool); ok {
		model.ConcurrentForwarding = types.BoolValue(v)
	}
	if v, ok := response["resolverRetries"].(float64); ok {
		model.ResolverRetries = types.Int64Value(int64(v))
	}
	if v, ok := response["resolverTimeout"].(float64); ok {
		model.ResolverTimeout = types.Int64Value(int64(v))
	}
	if v, ok := response["resolverConcurrency"].(float64); ok {
		model.ResolverConcurrency = types.Int64Value(int64(v))
	}
	if v, ok := response["resolverMaxStackCount"].(float64); ok {
		model.ResolverMaxStackCount = types.Int64Value(int64(v))
	}
	if v, ok := response["clientTimeout"].(float64); ok {
		model.ClientTimeout = types.Int64Value(int64(v))
	}
	if v, ok := response["udpPayloadSize"].(float64); ok {
		model.UdpPayloadSize = types.Int64Value(int64(v))
	}
	if v, ok := response["tcpReceiveTimeout"].(float64); ok {
		model.TcpReceiveTimeout = types.Int64Value(int64(v))
	}
	if v, ok := response["tcpSendTimeout"].(float64); ok {
		model.TcpSendTimeout = types.Int64Value(int64(v))
	}
	if v, ok := response["udpReceiveBufferSizeKB"].(float64); ok {
		model.UdpReceiveBufferSizeKb = types.Int64Value(int64(v))
	}
	if v, ok := response["udpSendBufferSizeKB"].(float64); ok {
		model.UdpSendBufferSizeKb = types.Int64Value(int64(v))
	}
	if v, ok := response["ipv6Mode"].(string); ok {
		model.Ipv6Mode = types.StringValue(v)
	}
	if v, ok := response["listenBacklog"].(float64); ok {
		model.ListenBacklog = types.Int64Value(int64(v))
	}
	if v, ok := response["defaultSoaRecordTtl"].(float64); ok {
		model.DefaultSoaRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["defaultNsRecordTtl"].(float64); ok {
		model.DefaultNsRecordTtl = types.Int64Value(int64(v))
	}
	if v, ok := response["minSoaRefresh"].(float64); ok {
		model.MinSoaRefresh = types.Int64Value(int64(v))
	}
	if v, ok := response["minSoaRetry"].(float64); ok {
		model.MinSoaRetry = types.Int64Value(int64(v))
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

	if !model.AllowTxtBlockingReport.IsNull() && !model.AllowTxtBlockingReport.IsUnknown() {
		params.Set("allowTxtBlockingReport", fmt.Sprintf("%t", model.AllowTxtBlockingReport.ValueBool()))
	}
	if !model.BlockingAnswerTtl.IsNull() && !model.BlockingAnswerTtl.IsUnknown() {
		params.Set("blockingAnswerTtl", fmt.Sprintf("%d", model.BlockingAnswerTtl.ValueInt64()))
	}
	if !model.BlockListUpdateIntervalHours.IsNull() && !model.BlockListUpdateIntervalHours.IsUnknown() {
		params.Set("blockListUpdateIntervalHours", fmt.Sprintf("%d", model.BlockListUpdateIntervalHours.ValueInt64()))
	}
	if !model.CachePrefetchEligibility.IsNull() && !model.CachePrefetchEligibility.IsUnknown() {
		params.Set("cachePrefetchEligibility", fmt.Sprintf("%d", model.CachePrefetchEligibility.ValueInt64()))
	}
	if !model.CachePrefetchTrigger.IsNull() && !model.CachePrefetchTrigger.IsUnknown() {
		params.Set("cachePrefetchTrigger", fmt.Sprintf("%d", model.CachePrefetchTrigger.ValueInt64()))
	}
	if !model.CachePrefetchSampleIntervalMinutes.IsNull() && !model.CachePrefetchSampleIntervalMinutes.IsUnknown() {
		params.Set("cachePrefetchSampleIntervalInMinutes", fmt.Sprintf("%d", model.CachePrefetchSampleIntervalMinutes.ValueInt64()))
	}
	if !model.CachePrefetchSampleEligibilityHitsPerHour.IsNull() && !model.CachePrefetchSampleEligibilityHitsPerHour.IsUnknown() {
		params.Set("cachePrefetchSampleEligibilityHitsPerHour", fmt.Sprintf("%d", model.CachePrefetchSampleEligibilityHitsPerHour.ValueInt64()))
	}
	if !model.CacheFailureRecordTtl.IsNull() && !model.CacheFailureRecordTtl.IsUnknown() {
		params.Set("cacheFailureRecordTtl", fmt.Sprintf("%d", model.CacheFailureRecordTtl.ValueInt64()))
	}
	if !model.SaveCache.IsNull() && !model.SaveCache.IsUnknown() {
		params.Set("saveCache", fmt.Sprintf("%t", model.SaveCache.ValueBool()))
	}
	if !model.ServeStaleTtl.IsNull() && !model.ServeStaleTtl.IsUnknown() {
		params.Set("serveStaleTtl", fmt.Sprintf("%d", model.ServeStaleTtl.ValueInt64()))
	}
	if !model.ServeStaleAnswerTtl.IsNull() && !model.ServeStaleAnswerTtl.IsUnknown() {
		params.Set("serveStaleAnswerTtl", fmt.Sprintf("%d", model.ServeStaleAnswerTtl.ValueInt64()))
	}
	if !model.ServeStaleMaxWaitTime.IsNull() && !model.ServeStaleMaxWaitTime.IsUnknown() {
		params.Set("serveStaleMaxWaitTime", fmt.Sprintf("%d", model.ServeStaleMaxWaitTime.ValueInt64()))
	}
	if !model.ServeStaleResetTtl.IsNull() && !model.ServeStaleResetTtl.IsUnknown() {
		params.Set("serveStaleResetTtl", fmt.Sprintf("%d", model.ServeStaleResetTtl.ValueInt64()))
	}
	if !model.ForwarderRetries.IsNull() && !model.ForwarderRetries.IsUnknown() {
		params.Set("forwarderRetries", fmt.Sprintf("%d", model.ForwarderRetries.ValueInt64()))
	}
	if !model.ForwarderTimeout.IsNull() && !model.ForwarderTimeout.IsUnknown() {
		params.Set("forwarderTimeout", fmt.Sprintf("%d", model.ForwarderTimeout.ValueInt64()))
	}
	if !model.ForwarderConcurrency.IsNull() && !model.ForwarderConcurrency.IsUnknown() {
		params.Set("forwarderConcurrency", fmt.Sprintf("%d", model.ForwarderConcurrency.ValueInt64()))
	}
	if !model.ConcurrentForwarding.IsNull() && !model.ConcurrentForwarding.IsUnknown() {
		params.Set("concurrentForwarding", fmt.Sprintf("%t", model.ConcurrentForwarding.ValueBool()))
	}
	if !model.ResolverRetries.IsNull() && !model.ResolverRetries.IsUnknown() {
		params.Set("resolverRetries", fmt.Sprintf("%d", model.ResolverRetries.ValueInt64()))
	}
	if !model.ResolverTimeout.IsNull() && !model.ResolverTimeout.IsUnknown() {
		params.Set("resolverTimeout", fmt.Sprintf("%d", model.ResolverTimeout.ValueInt64()))
	}
	if !model.ResolverConcurrency.IsNull() && !model.ResolverConcurrency.IsUnknown() {
		params.Set("resolverConcurrency", fmt.Sprintf("%d", model.ResolverConcurrency.ValueInt64()))
	}
	if !model.ResolverMaxStackCount.IsNull() && !model.ResolverMaxStackCount.IsUnknown() {
		params.Set("resolverMaxStackCount", fmt.Sprintf("%d", model.ResolverMaxStackCount.ValueInt64()))
	}
	if !model.ClientTimeout.IsNull() && !model.ClientTimeout.IsUnknown() {
		params.Set("clientTimeout", fmt.Sprintf("%d", model.ClientTimeout.ValueInt64()))
	}
	if !model.UdpPayloadSize.IsNull() && !model.UdpPayloadSize.IsUnknown() {
		params.Set("udpPayloadSize", fmt.Sprintf("%d", model.UdpPayloadSize.ValueInt64()))
	}
	if !model.TcpReceiveTimeout.IsNull() && !model.TcpReceiveTimeout.IsUnknown() {
		params.Set("tcpReceiveTimeout", fmt.Sprintf("%d", model.TcpReceiveTimeout.ValueInt64()))
	}
	if !model.TcpSendTimeout.IsNull() && !model.TcpSendTimeout.IsUnknown() {
		params.Set("tcpSendTimeout", fmt.Sprintf("%d", model.TcpSendTimeout.ValueInt64()))
	}
	if !model.UdpReceiveBufferSizeKb.IsNull() && !model.UdpReceiveBufferSizeKb.IsUnknown() {
		params.Set("udpReceiveBufferSizeKB", fmt.Sprintf("%d", model.UdpReceiveBufferSizeKb.ValueInt64()))
	}
	if !model.UdpSendBufferSizeKb.IsNull() && !model.UdpSendBufferSizeKb.IsUnknown() {
		params.Set("udpSendBufferSizeKB", fmt.Sprintf("%d", model.UdpSendBufferSizeKb.ValueInt64()))
	}
	if !model.Ipv6Mode.IsNull() && !model.Ipv6Mode.IsUnknown() {
		params.Set("ipv6Mode", model.Ipv6Mode.ValueString())
	}
	if !model.ListenBacklog.IsNull() && !model.ListenBacklog.IsUnknown() {
		params.Set("listenBacklog", fmt.Sprintf("%d", model.ListenBacklog.ValueInt64()))
	}
	if !model.DefaultSoaRecordTtl.IsNull() && !model.DefaultSoaRecordTtl.IsUnknown() {
		params.Set("defaultSoaRecordTtl", fmt.Sprintf("%d", model.DefaultSoaRecordTtl.ValueInt64()))
	}
	if !model.DefaultNsRecordTtl.IsNull() && !model.DefaultNsRecordTtl.IsUnknown() {
		params.Set("defaultNsRecordTtl", fmt.Sprintf("%d", model.DefaultNsRecordTtl.ValueInt64()))
	}
	if !model.MinSoaRefresh.IsNull() && !model.MinSoaRefresh.IsUnknown() {
		params.Set("minSoaRefresh", fmt.Sprintf("%d", model.MinSoaRefresh.ValueInt64()))
	}
	if !model.MinSoaRetry.IsNull() && !model.MinSoaRetry.IsUnknown() {
		params.Set("minSoaRetry", fmt.Sprintf("%d", model.MinSoaRetry.ValueInt64()))
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
