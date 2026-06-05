package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &dnsSettingsDataSource{}

func NewDNSSettingsDataSource() datasource.DataSource {
	return &dnsSettingsDataSource{}
}

type dnsSettingsDataSource struct {
	client *client.Client
}

type dnsSettingsDataSourceModel struct {
	DnsServerDomain   types.String `tfsdk:"dns_server_domain"`
	DnssecValidation  types.Bool   `tfsdk:"dnssec_validation"`
	Recursion         types.String `tfsdk:"recursion"`
	PreferIPv6        types.Bool   `tfsdk:"prefer_ipv6"`
	EnableBlocking    types.Bool   `tfsdk:"enable_blocking"`
	BlockingType      types.String `tfsdk:"blocking_type"`
	Forwarders        types.List   `tfsdk:"forwarders"`
	ForwarderProtocol types.String `tfsdk:"forwarder_protocol"`
	EnableLogging                             types.Bool   `tfsdk:"enable_logging"`
	LogQueries                                types.Bool   `tfsdk:"log_queries"`
	AllowTxtBlockingReport                    types.Bool   `tfsdk:"allow_txt_blocking_report"`
	BlockingAnswerTtl                         types.Int64  `tfsdk:"blocking_answer_ttl"`
	BlockListUpdateIntervalHours              types.Int64  `tfsdk:"block_list_update_interval_hours"`
	CachePrefetchEligibility                  types.Int64  `tfsdk:"cache_prefetch_eligibility"`
	CachePrefetchTrigger                      types.Int64  `tfsdk:"cache_prefetch_trigger"`
	CachePrefetchSampleIntervalMinutes        types.Int64  `tfsdk:"cache_prefetch_sample_interval_minutes"`
	CachePrefetchSampleEligibilityHitsPerHour types.Int64  `tfsdk:"cache_prefetch_sample_eligibility_hits_per_hour"`
	CacheFailureRecordTtl                     types.Int64  `tfsdk:"cache_failure_record_ttl"`
	SaveCache                                 types.Bool   `tfsdk:"save_cache"`
	ServeStaleTtl                             types.Int64  `tfsdk:"serve_stale_ttl"`
	ServeStaleAnswerTtl                       types.Int64  `tfsdk:"serve_stale_answer_ttl"`
	ServeStaleMaxWaitTime                     types.Int64  `tfsdk:"serve_stale_max_wait_time"`
	ServeStaleResetTtl                        types.Int64  `tfsdk:"serve_stale_reset_ttl"`
	ForwarderRetries                          types.Int64  `tfsdk:"forwarder_retries"`
	ForwarderTimeout                          types.Int64  `tfsdk:"forwarder_timeout"`
	ForwarderConcurrency                      types.Int64  `tfsdk:"forwarder_concurrency"`
	ConcurrentForwarding                      types.Bool   `tfsdk:"concurrent_forwarding"`
	ResolverRetries                           types.Int64  `tfsdk:"resolver_retries"`
	ResolverTimeout                           types.Int64  `tfsdk:"resolver_timeout"`
	ResolverConcurrency                       types.Int64  `tfsdk:"resolver_concurrency"`
	ResolverMaxStackCount                     types.Int64  `tfsdk:"resolver_max_stack_count"`
	ClientTimeout                             types.Int64  `tfsdk:"client_timeout"`
	UdpPayloadSize                            types.Int64  `tfsdk:"udp_payload_size"`
	TcpReceiveTimeout                         types.Int64  `tfsdk:"tcp_receive_timeout"`
	TcpSendTimeout                            types.Int64  `tfsdk:"tcp_send_timeout"`
	UdpReceiveBufferSizeKb                    types.Int64  `tfsdk:"udp_receive_buffer_size_kb"`
	UdpSendBufferSizeKb                       types.Int64  `tfsdk:"udp_send_buffer_size_kb"`
	Ipv6Mode                                  types.String `tfsdk:"ipv6_mode"`
	ListenBacklog                             types.Int64  `tfsdk:"listen_backlog"`
	DefaultSoaRecordTtl                       types.Int64  `tfsdk:"default_soa_record_ttl"`
	DefaultNsRecordTtl                        types.Int64  `tfsdk:"default_ns_record_ttl"`
	MinSoaRefresh                             types.Int64  `tfsdk:"min_soa_refresh"`
	MinSoaRetry                               types.Int64  `tfsdk:"min_soa_retry"`
}

func (d *dnsSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
}

func (d *dnsSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the DNS server settings from a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"dns_server_domain": schema.StringAttribute{
				Description: "The primary domain name used by the DNS server.",
				Computed:    true,
			},
			"dnssec_validation": schema.BoolAttribute{
				Description: "Whether DNSSEC validation is enabled.",
				Computed:    true,
			},
			"recursion": schema.StringAttribute{
				Description: "The recursion policy.",
				Computed:    true,
			},
			"prefer_ipv6": schema.BoolAttribute{
				Description: "Whether IPv6 is preferred for DNS resolution.",
				Computed:    true,
			},
			"enable_blocking": schema.BoolAttribute{
				Description: "Whether DNS-level blocking is enabled.",
				Computed:    true,
			},
			"blocking_type": schema.StringAttribute{
				Description: "How blocked queries are answered.",
				Computed:    true,
			},
			"forwarders": schema.ListAttribute{
				Description: "List of forwarder DNS server addresses.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"forwarder_protocol": schema.StringAttribute{
				Description: "Protocol used for DNS forwarding.",
				Computed:    true,
			},
			"enable_logging": schema.BoolAttribute{
				Description: "Whether DNS query logging is enabled.",
				Computed:    true,
			},
			"log_queries": schema.BoolAttribute{
				Description: "Whether all DNS queries are logged.",
				Computed:    true,
			},
			"allow_txt_blocking_report": schema.BoolAttribute{
				Description: "Whether TXT record blocking report queries are allowed.",
				Computed:    true,
			},
			"blocking_answer_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for blocked DNS responses.",
				Computed:    true,
			},
			"block_list_update_interval_hours": schema.Int64Attribute{
				Description: "Interval in hours between block list updates.",
				Computed:    true,
			},
			"cache_prefetch_eligibility": schema.Int64Attribute{
				Description: "Minimum number of hits for a record to be eligible for cache prefetch.",
				Computed:    true,
			},
			"cache_prefetch_trigger": schema.Int64Attribute{
				Description: "Number of hits to trigger a cache prefetch.",
				Computed:    true,
			},
			"cache_prefetch_sample_interval_minutes": schema.Int64Attribute{
				Description: "Interval in minutes between cache prefetch sampling.",
				Computed:    true,
			},
			"cache_prefetch_sample_eligibility_hits_per_hour": schema.Int64Attribute{
				Description: "Minimum hits per hour for a record to be eligible for prefetch sampling.",
				Computed:    true,
			},
			"cache_failure_record_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for cached failure records.",
				Computed:    true,
			},
			"save_cache": schema.BoolAttribute{
				Description: "Whether DNS cache is saved to disk on shutdown.",
				Computed:    true,
			},
			"serve_stale_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for serve stale records.",
				Computed:    true,
			},
			"serve_stale_answer_ttl": schema.Int64Attribute{
				Description: "TTL in seconds used in stale answers.",
				Computed:    true,
			},
			"serve_stale_max_wait_time": schema.Int64Attribute{
				Description: "Maximum wait time in milliseconds before serving stale.",
				Computed:    true,
			},
			"serve_stale_reset_ttl": schema.Int64Attribute{
				Description: "TTL in seconds to reset serve stale timer.",
				Computed:    true,
			},
			"forwarder_retries": schema.Int64Attribute{
				Description: "Number of retries for forwarder queries.",
				Computed:    true,
			},
			"forwarder_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for forwarder queries.",
				Computed:    true,
			},
			"forwarder_concurrency": schema.Int64Attribute{
				Description: "Number of concurrent forwarder queries.",
				Computed:    true,
			},
			"concurrent_forwarding": schema.BoolAttribute{
				Description: "Whether concurrent forwarding to all configured forwarders is enabled.",
				Computed:    true,
			},
			"resolver_retries": schema.Int64Attribute{
				Description: "Number of retries for resolver queries.",
				Computed:    true,
			},
			"resolver_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for resolver queries.",
				Computed:    true,
			},
			"resolver_concurrency": schema.Int64Attribute{
				Description: "Number of concurrent resolver queries.",
				Computed:    true,
			},
			"resolver_max_stack_count": schema.Int64Attribute{
				Description: "Maximum number of resolver stack entries.",
				Computed:    true,
			},
			"client_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for client connections.",
				Computed:    true,
			},
			"udp_payload_size": schema.Int64Attribute{
				Description: "Maximum UDP payload size in bytes.",
				Computed:    true,
			},
			"tcp_receive_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for TCP receive operations.",
				Computed:    true,
			},
			"tcp_send_timeout": schema.Int64Attribute{
				Description: "Timeout in milliseconds for TCP send operations.",
				Computed:    true,
			},
			"udp_receive_buffer_size_kb": schema.Int64Attribute{
				Description: "UDP receive buffer size in kilobytes.",
				Computed:    true,
			},
			"udp_send_buffer_size_kb": schema.Int64Attribute{
				Description: "UDP send buffer size in kilobytes.",
				Computed:    true,
			},
			"ipv6_mode": schema.StringAttribute{
				Description: "IPv6 mode.",
				Computed:    true,
			},
			"listen_backlog": schema.Int64Attribute{
				Description: "TCP listen backlog size.",
				Computed:    true,
			},
			"default_soa_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for SOA records.",
				Computed:    true,
			},
			"default_ns_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for NS records.",
				Computed:    true,
			},
			"min_soa_refresh": schema.Int64Attribute{
				Description: "Minimum SOA refresh interval in seconds.",
				Computed:    true,
			},
			"min_soa_retry": schema.Int64Attribute{
				Description: "Minimum SOA retry interval in seconds.",
				Computed:    true,
			},
		},
	}
}

func (d *dnsSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *dnsSettingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading DNS settings data source")

	response, err := d.client.GetDNSSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DNS Settings",
			fmt.Sprintf("Could not read DNS settings: %s", err),
		)
		return
	}

	var state dnsSettingsDataSourceModel

	if v, ok := response["dnsServerDomain"].(string); ok {
		state.DnsServerDomain = types.StringValue(v)
	} else {
		state.DnsServerDomain = types.StringNull()
	}

	if v, ok := response["dnssecValidation"].(bool); ok {
		state.DnssecValidation = types.BoolValue(v)
	} else {
		state.DnssecValidation = types.BoolNull()
	}

	if v, ok := response["recursion"].(string); ok {
		state.Recursion = types.StringValue(v)
	} else {
		state.Recursion = types.StringNull()
	}

	if v, ok := response["preferIPv6"].(bool); ok {
		state.PreferIPv6 = types.BoolValue(v)
	} else {
		state.PreferIPv6 = types.BoolNull()
	}

	if v, ok := response["enableBlocking"].(bool); ok {
		state.EnableBlocking = types.BoolValue(v)
	} else {
		state.EnableBlocking = types.BoolNull()
	}

	if v, ok := response["blockingType"].(string); ok {
		state.BlockingType = types.StringValue(v)
	} else {
		state.BlockingType = types.StringNull()
	}

	if v, ok := response["forwarderProtocol"].(string); ok {
		state.ForwarderProtocol = types.StringValue(v)
	} else {
		state.ForwarderProtocol = types.StringNull()
	}

	if v, ok := response["enableLogging"].(bool); ok {
		state.EnableLogging = types.BoolValue(v)
	} else {
		state.EnableLogging = types.BoolNull()
	}

	if v, ok := response["logQueries"].(bool); ok {
		state.LogQueries = types.BoolValue(v)
	} else {
		state.LogQueries = types.BoolNull()
	}

	if v, ok := response["allowTxtBlockingReport"].(bool); ok {
		state.AllowTxtBlockingReport = types.BoolValue(v)
	} else {
		state.AllowTxtBlockingReport = types.BoolNull()
	}

	if v, ok := response["blockingAnswerTtl"].(float64); ok {
		state.BlockingAnswerTtl = types.Int64Value(int64(v))
	} else {
		state.BlockingAnswerTtl = types.Int64Null()
	}

	if v, ok := response["blockListUpdateIntervalHours"].(float64); ok {
		state.BlockListUpdateIntervalHours = types.Int64Value(int64(v))
	} else {
		state.BlockListUpdateIntervalHours = types.Int64Null()
	}

	if v, ok := response["cachePrefetchEligibility"].(float64); ok {
		state.CachePrefetchEligibility = types.Int64Value(int64(v))
	} else {
		state.CachePrefetchEligibility = types.Int64Null()
	}

	if v, ok := response["cachePrefetchTrigger"].(float64); ok {
		state.CachePrefetchTrigger = types.Int64Value(int64(v))
	} else {
		state.CachePrefetchTrigger = types.Int64Null()
	}

	if v, ok := response["cachePrefetchSampleIntervalInMinutes"].(float64); ok {
		state.CachePrefetchSampleIntervalMinutes = types.Int64Value(int64(v))
	} else {
		state.CachePrefetchSampleIntervalMinutes = types.Int64Null()
	}

	if v, ok := response["cachePrefetchSampleEligibilityHitsPerHour"].(float64); ok {
		state.CachePrefetchSampleEligibilityHitsPerHour = types.Int64Value(int64(v))
	} else {
		state.CachePrefetchSampleEligibilityHitsPerHour = types.Int64Null()
	}

	if v, ok := response["cacheFailureRecordTtl"].(float64); ok {
		state.CacheFailureRecordTtl = types.Int64Value(int64(v))
	} else {
		state.CacheFailureRecordTtl = types.Int64Null()
	}

	if v, ok := response["saveCache"].(bool); ok {
		state.SaveCache = types.BoolValue(v)
	} else {
		state.SaveCache = types.BoolNull()
	}

	if v, ok := response["serveStaleTtl"].(float64); ok {
		state.ServeStaleTtl = types.Int64Value(int64(v))
	} else {
		state.ServeStaleTtl = types.Int64Null()
	}

	if v, ok := response["serveStaleAnswerTtl"].(float64); ok {
		state.ServeStaleAnswerTtl = types.Int64Value(int64(v))
	} else {
		state.ServeStaleAnswerTtl = types.Int64Null()
	}

	if v, ok := response["serveStaleMaxWaitTime"].(float64); ok {
		state.ServeStaleMaxWaitTime = types.Int64Value(int64(v))
	} else {
		state.ServeStaleMaxWaitTime = types.Int64Null()
	}

	if v, ok := response["serveStaleResetTtl"].(float64); ok {
		state.ServeStaleResetTtl = types.Int64Value(int64(v))
	} else {
		state.ServeStaleResetTtl = types.Int64Null()
	}

	if v, ok := response["forwarderRetries"].(float64); ok {
		state.ForwarderRetries = types.Int64Value(int64(v))
	} else {
		state.ForwarderRetries = types.Int64Null()
	}

	if v, ok := response["forwarderTimeout"].(float64); ok {
		state.ForwarderTimeout = types.Int64Value(int64(v))
	} else {
		state.ForwarderTimeout = types.Int64Null()
	}

	if v, ok := response["forwarderConcurrency"].(float64); ok {
		state.ForwarderConcurrency = types.Int64Value(int64(v))
	} else {
		state.ForwarderConcurrency = types.Int64Null()
	}

	if v, ok := response["concurrentForwarding"].(bool); ok {
		state.ConcurrentForwarding = types.BoolValue(v)
	} else {
		state.ConcurrentForwarding = types.BoolNull()
	}

	if v, ok := response["resolverRetries"].(float64); ok {
		state.ResolverRetries = types.Int64Value(int64(v))
	} else {
		state.ResolverRetries = types.Int64Null()
	}

	if v, ok := response["resolverTimeout"].(float64); ok {
		state.ResolverTimeout = types.Int64Value(int64(v))
	} else {
		state.ResolverTimeout = types.Int64Null()
	}

	if v, ok := response["resolverConcurrency"].(float64); ok {
		state.ResolverConcurrency = types.Int64Value(int64(v))
	} else {
		state.ResolverConcurrency = types.Int64Null()
	}

	if v, ok := response["resolverMaxStackCount"].(float64); ok {
		state.ResolverMaxStackCount = types.Int64Value(int64(v))
	} else {
		state.ResolverMaxStackCount = types.Int64Null()
	}

	if v, ok := response["clientTimeout"].(float64); ok {
		state.ClientTimeout = types.Int64Value(int64(v))
	} else {
		state.ClientTimeout = types.Int64Null()
	}

	if v, ok := response["udpPayloadSize"].(float64); ok {
		state.UdpPayloadSize = types.Int64Value(int64(v))
	} else {
		state.UdpPayloadSize = types.Int64Null()
	}

	if v, ok := response["tcpReceiveTimeout"].(float64); ok {
		state.TcpReceiveTimeout = types.Int64Value(int64(v))
	} else {
		state.TcpReceiveTimeout = types.Int64Null()
	}

	if v, ok := response["tcpSendTimeout"].(float64); ok {
		state.TcpSendTimeout = types.Int64Value(int64(v))
	} else {
		state.TcpSendTimeout = types.Int64Null()
	}

	if v, ok := response["udpReceiveBufferSizeKB"].(float64); ok {
		state.UdpReceiveBufferSizeKb = types.Int64Value(int64(v))
	} else {
		state.UdpReceiveBufferSizeKb = types.Int64Null()
	}

	if v, ok := response["udpSendBufferSizeKB"].(float64); ok {
		state.UdpSendBufferSizeKb = types.Int64Value(int64(v))
	} else {
		state.UdpSendBufferSizeKb = types.Int64Null()
	}

	if v, ok := response["ipv6Mode"].(string); ok {
		state.Ipv6Mode = types.StringValue(v)
	} else {
		state.Ipv6Mode = types.StringNull()
	}

	if v, ok := response["listenBacklog"].(float64); ok {
		state.ListenBacklog = types.Int64Value(int64(v))
	} else {
		state.ListenBacklog = types.Int64Null()
	}

	if v, ok := response["defaultSoaRecordTtl"].(float64); ok {
		state.DefaultSoaRecordTtl = types.Int64Value(int64(v))
	} else {
		state.DefaultSoaRecordTtl = types.Int64Null()
	}

	if v, ok := response["defaultNsRecordTtl"].(float64); ok {
		state.DefaultNsRecordTtl = types.Int64Value(int64(v))
	} else {
		state.DefaultNsRecordTtl = types.Int64Null()
	}

	if v, ok := response["minSoaRefresh"].(float64); ok {
		state.MinSoaRefresh = types.Int64Value(int64(v))
	} else {
		state.MinSoaRefresh = types.Int64Null()
	}

	if v, ok := response["minSoaRetry"].(float64); ok {
		state.MinSoaRetry = types.Int64Value(int64(v))
	} else {
		state.MinSoaRetry = types.Int64Null()
	}

	if v, ok := response["forwarders"].([]interface{}); ok {
		fwds := make([]string, len(v))
		for i, f := range v {
			fwds[i], _ = f.(string)
		}
		listVal, diags := types.ListValueFrom(ctx, types.StringType, fwds)
		resp.Diagnostics.Append(diags...)
		state.Forwarders = listVal
	} else {
		state.Forwarders = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
