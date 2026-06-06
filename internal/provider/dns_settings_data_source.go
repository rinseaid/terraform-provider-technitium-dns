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
	DnsServerDomain                           types.String `tfsdk:"dns_server_domain"`
	DefaultRecordTtl                          types.Int64  `tfsdk:"default_record_ttl"`
	DnssecValidation                          types.Bool   `tfsdk:"dnssec_validation"`
	Recursion                                 types.String `tfsdk:"recursion"`
	QnameMinimization                         types.Bool   `tfsdk:"qname_minimization"`
	RandomizeName                             types.Bool   `tfsdk:"randomize_name"`
	PreferIPv6                                types.Bool   `tfsdk:"prefer_ipv6"`
	EnableBlocking                            types.Bool   `tfsdk:"enable_blocking"`
	BlockingType                              types.String `tfsdk:"blocking_type"`
	BlockListUrls                             types.List   `tfsdk:"block_list_urls"`
	Forwarders                                types.List   `tfsdk:"forwarders"`
	ForwarderProtocol                         types.String `tfsdk:"forwarder_protocol"`
	EnableLogging                             types.Bool   `tfsdk:"enable_logging"`
	MaxLogFileDays                            types.Int64  `tfsdk:"max_log_file_days"`
	LogQueries                                types.Bool   `tfsdk:"log_queries"`
	AllowTxtBlockingReport                    types.Bool   `tfsdk:"allow_txt_blocking_report"`
	BlockingAnswerTtl                         types.Int64  `tfsdk:"blocking_answer_ttl"`
	BlockListUpdateIntervalHours              types.Int64  `tfsdk:"block_list_update_interval_hours"`
	CachePrefetchEligibility                  types.Int64  `tfsdk:"cache_prefetch_eligibility"`
	CachePrefetchTrigger                      types.Int64  `tfsdk:"cache_prefetch_trigger"`
	CachePrefetchSampleIntervalMinutes        types.Int64  `tfsdk:"cache_prefetch_sample_interval_minutes"`
	CachePrefetchSampleEligibilityHitsPerHour types.Int64  `tfsdk:"cache_prefetch_sample_eligibility_hits_per_hour"`
	CacheMaximumEntries                       types.Int64  `tfsdk:"cache_maximum_entries"`
	CacheMinimumRecordTtl                     types.Int64  `tfsdk:"cache_minimum_record_ttl"`
	CacheMaximumRecordTtl                     types.Int64  `tfsdk:"cache_maximum_record_ttl"`
	CacheNegativeRecordTtl                    types.Int64  `tfsdk:"cache_negative_record_ttl"`
	CacheFailureRecordTtl                     types.Int64  `tfsdk:"cache_failure_record_ttl"`
	SaveCache                                 types.Bool   `tfsdk:"save_cache"`
	ServeStale                                types.Bool   `tfsdk:"serve_stale"`
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
	EnableDnsOverUdpProxy                     types.Bool   `tfsdk:"enable_dns_over_udp_proxy"`
	EnableDnsOverTcpProxy                     types.Bool   `tfsdk:"enable_dns_over_tcp_proxy"`
	EnableDnsOverHttp                         types.Bool   `tfsdk:"enable_dns_over_http"`
	EnableDnsOverTls                          types.Bool   `tfsdk:"enable_dns_over_tls"`
	EnableDnsOverHttps                        types.Bool   `tfsdk:"enable_dns_over_https"`
	EnableDnsOverHttp3                        types.Bool   `tfsdk:"enable_dns_over_http3"`
	EnableDnsOverQuic                         types.Bool   `tfsdk:"enable_dns_over_quic"`
	DnsOverUdpProxyPort                       types.Int64  `tfsdk:"dns_over_udp_proxy_port"`
	DnsOverTcpProxyPort                       types.Int64  `tfsdk:"dns_over_tcp_proxy_port"`
	DnsOverHttpPort                           types.Int64  `tfsdk:"dns_over_http_port"`
	DnsOverTlsPort                            types.Int64  `tfsdk:"dns_over_tls_port"`
	DnsOverHttpsPort                          types.Int64  `tfsdk:"dns_over_https_port"`
	DnsOverQuicPort                           types.Int64  `tfsdk:"dns_over_quic_port"`
	DnsTlsCertificatePath                     types.String `tfsdk:"dns_tls_certificate_path"`
	DnsTlsCertificatePassword                 types.String `tfsdk:"dns_tls_certificate_password"`
	WebServiceHttpPort                        types.Int64  `tfsdk:"web_service_http_port"`
	WebServiceTlsPort                         types.Int64  `tfsdk:"web_service_tls_port"`
	WebServiceEnableTls                       types.Bool   `tfsdk:"web_service_enable_tls"`
	WebServiceEnableHttp3                     types.Bool   `tfsdk:"web_service_enable_http3"`
	WebServiceHttpToTlsRedirect               types.Bool   `tfsdk:"web_service_http_to_tls_redirect"`
	WebServiceUseSelfSignedTlsCertificate     types.Bool   `tfsdk:"web_service_use_self_signed_tls_certificate"`
	WebServiceTlsCertificatePath              types.String `tfsdk:"web_service_tls_certificate_path"`
	WebServiceTlsCertificatePassword          types.String `tfsdk:"web_service_tls_certificate_password"`
	WebServiceRealIpHeader                    types.String `tfsdk:"web_service_real_ip_header"`
	DnsOverHttpRealIpHeader                   types.String `tfsdk:"dns_over_http_real_ip_header"`
	ServerProxyType                           types.String `tfsdk:"server_proxy_type"`
	ServerProxyAddress                        types.String `tfsdk:"server_proxy_address"`
	ServerProxyPort                           types.Int64  `tfsdk:"server_proxy_port"`
	ServerProxyUsername                       types.String `tfsdk:"server_proxy_username"`
	ServerProxyPassword                       types.String `tfsdk:"server_proxy_password"`
	ServerProxyBypass                         types.String `tfsdk:"server_proxy_bypass"`
	EdnsClientSubnet                          types.Bool   `tfsdk:"edns_client_subnet"`
	EdnsClientSubnetIpv4PrefixLength          types.Int64  `tfsdk:"edns_client_subnet_ipv4_prefix_length"`
	EdnsClientSubnetIpv6PrefixLength          types.Int64  `tfsdk:"edns_client_subnet_ipv6_prefix_length"`
	EdnsClientSubnetIpv4Override              types.String `tfsdk:"edns_client_subnet_ipv4_override"`
	EdnsClientSubnetIpv6Override              types.String `tfsdk:"edns_client_subnet_ipv6_override"`
	DefaultResponsiblePerson                  types.String `tfsdk:"default_responsible_person"`
	UseSoaSerialDateScheme                    types.Bool   `tfsdk:"use_soa_serial_date_scheme"`
	DnsAppsEnableAutomaticUpdate              types.Bool   `tfsdk:"dns_apps_enable_automatic_update"`
	EnableUdpSocketPool                       types.Bool   `tfsdk:"enable_udp_socket_pool"`
	QuicIdleTimeout                           types.Int64  `tfsdk:"quic_idle_timeout"`
	QuicMaxInboundStreams                     types.Int64  `tfsdk:"quic_max_inbound_streams"`
	LoggingType                               types.String `tfsdk:"logging_type"`
	IgnoreResolverLogs                        types.Bool   `tfsdk:"ignore_resolver_logs"`
	UseLocalTime                              types.Bool   `tfsdk:"use_local_time"`
	LogFolder                                 types.String `tfsdk:"log_folder"`
	EnableInMemoryStats                       types.Bool   `tfsdk:"enable_in_memory_stats"`
	MaxStatFileDays                           types.Int64  `tfsdk:"max_stat_file_days"`
	MaxConcurrentResolutionsPerCore           types.Int64  `tfsdk:"max_concurrent_resolutions_per_core"`
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
			"default_record_ttl": schema.Int64Attribute{
				Description: "Default TTL in seconds for new records.",
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
			"qname_minimization": schema.BoolAttribute{
				Description: "Whether QNAME minimization is enabled for recursive queries.",
				Computed:    true,
			},
			"randomize_name": schema.BoolAttribute{
				Description: "Whether query name casing is randomized for cache poisoning protection.",
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
			"block_list_urls": schema.ListAttribute{
				Description: "List of block list URLs.",
				Computed:    true,
				ElementType: types.StringType,
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
			"max_log_file_days": schema.Int64Attribute{
				Description: "Maximum number of days to keep log files.",
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
			"cache_maximum_entries": schema.Int64Attribute{
				Description: "Maximum number of entries in the DNS cache.",
				Computed:    true,
			},
			"cache_minimum_record_ttl": schema.Int64Attribute{
				Description: "Minimum TTL in seconds for cached records.",
				Computed:    true,
			},
			"cache_maximum_record_ttl": schema.Int64Attribute{
				Description: "Maximum TTL in seconds for cached records.",
				Computed:    true,
			},
			"cache_negative_record_ttl": schema.Int64Attribute{
				Description: "TTL in seconds for negative cached records.",
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
			"serve_stale": schema.BoolAttribute{
				Description: "Whether stale cached records are served when upstream is unavailable.",
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
			"enable_dns_over_udp_proxy": schema.BoolAttribute{
				Description: "Whether DNS-over-UDP proxy protocol is enabled.",
				Computed:    true,
			},
			"enable_dns_over_tcp_proxy": schema.BoolAttribute{
				Description: "Whether DNS-over-TCP proxy protocol is enabled.",
				Computed:    true,
			},
			"enable_dns_over_http": schema.BoolAttribute{
				Description: "Whether DNS-over-HTTP is enabled.",
				Computed:    true,
			},
			"enable_dns_over_tls": schema.BoolAttribute{
				Description: "Whether DNS-over-TLS is enabled.",
				Computed:    true,
			},
			"enable_dns_over_https": schema.BoolAttribute{
				Description: "Whether DNS-over-HTTPS is enabled.",
				Computed:    true,
			},
			"enable_dns_over_http3": schema.BoolAttribute{
				Description: "Whether DNS-over-HTTP/3 is enabled.",
				Computed:    true,
			},
			"enable_dns_over_quic": schema.BoolAttribute{
				Description: "Whether DNS-over-QUIC is enabled.",
				Computed:    true,
			},
			"dns_over_udp_proxy_port": schema.Int64Attribute{
				Description: "Port for DNS-over-UDP proxy protocol.",
				Computed:    true,
			},
			"dns_over_tcp_proxy_port": schema.Int64Attribute{
				Description: "Port for DNS-over-TCP proxy protocol.",
				Computed:    true,
			},
			"dns_over_http_port": schema.Int64Attribute{
				Description: "Port for DNS-over-HTTP.",
				Computed:    true,
			},
			"dns_over_tls_port": schema.Int64Attribute{
				Description: "Port for DNS-over-TLS.",
				Computed:    true,
			},
			"dns_over_https_port": schema.Int64Attribute{
				Description: "Port for DNS-over-HTTPS.",
				Computed:    true,
			},
			"dns_over_quic_port": schema.Int64Attribute{
				Description: "Port for DNS-over-QUIC.",
				Computed:    true,
			},
			"dns_tls_certificate_path": schema.StringAttribute{
				Description: "File path to the TLS certificate for DNS-over-TLS and DNS-over-HTTPS.",
				Computed:    true,
			},
			"dns_tls_certificate_password": schema.StringAttribute{
				Description: "Password for the DNS TLS certificate.",
				Computed:    true,
				Sensitive:   true,
			},
			"web_service_http_port": schema.Int64Attribute{
				Description: "HTTP port for the web service.",
				Computed:    true,
			},
			"web_service_tls_port": schema.Int64Attribute{
				Description: "TLS port for the web service.",
				Computed:    true,
			},
			"web_service_enable_tls": schema.BoolAttribute{
				Description: "Whether TLS is enabled for the web service.",
				Computed:    true,
			},
			"web_service_enable_http3": schema.BoolAttribute{
				Description: "Whether HTTP/3 is enabled for the web service.",
				Computed:    true,
			},
			"web_service_http_to_tls_redirect": schema.BoolAttribute{
				Description: "Whether HTTP to TLS redirect is enabled for the web service.",
				Computed:    true,
			},
			"web_service_use_self_signed_tls_certificate": schema.BoolAttribute{
				Description: "Whether a self-signed TLS certificate is used for the web service.",
				Computed:    true,
			},
			"web_service_tls_certificate_path": schema.StringAttribute{
				Description: "File path to the TLS certificate for the web service.",
				Computed:    true,
			},
			"web_service_tls_certificate_password": schema.StringAttribute{
				Description: "Password for the web service TLS certificate.",
				Computed:    true,
				Sensitive:   true,
			},
			"web_service_real_ip_header": schema.StringAttribute{
				Description: "HTTP header name for real IP detection in the web service.",
				Computed:    true,
			},
			"dns_over_http_real_ip_header": schema.StringAttribute{
				Description: "HTTP header name for real IP detection in DNS-over-HTTP.",
				Computed:    true,
			},
			"server_proxy_type": schema.StringAttribute{
				Description: "Proxy type for the DNS server.",
				Computed:    true,
			},
			"server_proxy_address": schema.StringAttribute{
				Description: "Proxy server address.",
				Computed:    true,
			},
			"server_proxy_port": schema.Int64Attribute{
				Description: "Proxy server port.",
				Computed:    true,
			},
			"server_proxy_username": schema.StringAttribute{
				Description: "Proxy server username.",
				Computed:    true,
				Sensitive:   true,
			},
			"server_proxy_password": schema.StringAttribute{
				Description: "Proxy server password.",
				Computed:    true,
				Sensitive:   true,
			},
			"server_proxy_bypass": schema.StringAttribute{
				Description: "Proxy bypass list.",
				Computed:    true,
			},
			"edns_client_subnet": schema.BoolAttribute{
				Description: "Whether EDNS Client Subnet is enabled.",
				Computed:    true,
			},
			"edns_client_subnet_ipv4_prefix_length": schema.Int64Attribute{
				Description: "EDNS Client Subnet IPv4 prefix length.",
				Computed:    true,
			},
			"edns_client_subnet_ipv6_prefix_length": schema.Int64Attribute{
				Description: "EDNS Client Subnet IPv6 prefix length.",
				Computed:    true,
			},
			"edns_client_subnet_ipv4_override": schema.StringAttribute{
				Description: "EDNS Client Subnet IPv4 override address.",
				Computed:    true,
			},
			"edns_client_subnet_ipv6_override": schema.StringAttribute{
				Description: "EDNS Client Subnet IPv6 override address.",
				Computed:    true,
			},
			"default_responsible_person": schema.StringAttribute{
				Description: "Default responsible person email for SOA records.",
				Computed:    true,
			},
			"use_soa_serial_date_scheme": schema.BoolAttribute{
				Description: "Whether date-based SOA serial number scheme is used.",
				Computed:    true,
			},
			"dns_apps_enable_automatic_update": schema.BoolAttribute{
				Description: "Whether automatic updates for DNS apps are enabled.",
				Computed:    true,
			},
			"enable_udp_socket_pool": schema.BoolAttribute{
				Description: "Whether UDP socket pooling is enabled.",
				Computed:    true,
			},
			"quic_idle_timeout": schema.Int64Attribute{
				Description: "QUIC idle timeout in milliseconds.",
				Computed:    true,
			},
			"quic_max_inbound_streams": schema.Int64Attribute{
				Description: "Maximum number of inbound QUIC streams.",
				Computed:    true,
			},
			"logging_type": schema.StringAttribute{
				Description: "Logging type.",
				Computed:    true,
			},
			"ignore_resolver_logs": schema.BoolAttribute{
				Description: "Whether resolver log entries are ignored.",
				Computed:    true,
			},
			"use_local_time": schema.BoolAttribute{
				Description: "Whether local time is used in logs.",
				Computed:    true,
			},
			"log_folder": schema.StringAttribute{
				Description: "Path to the log folder.",
				Computed:    true,
			},
			"enable_in_memory_stats": schema.BoolAttribute{
				Description: "Whether in-memory statistics are enabled.",
				Computed:    true,
			},
			"max_stat_file_days": schema.Int64Attribute{
				Description: "Number of days to retain statistics files.",
				Computed:    true,
			},
			"max_concurrent_resolutions_per_core": schema.Int64Attribute{
				Description: "Maximum concurrent DNS resolutions per CPU core.",
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

	response, err := d.client.GetDNSSettings(ctx)
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

	if v, ok := response["defaultRecordTtl"].(float64); ok {
		state.DefaultRecordTtl = types.Int64Value(int64(v))
	} else {
		state.DefaultRecordTtl = types.Int64Null()
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

	if v, ok := response["qnameMinimization"].(bool); ok {
		state.QnameMinimization = types.BoolValue(v)
	} else {
		state.QnameMinimization = types.BoolNull()
	}

	if v, ok := response["randomizeName"].(bool); ok {
		state.RandomizeName = types.BoolValue(v)
	} else {
		state.RandomizeName = types.BoolNull()
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

	if v, ok := response["blockListUrls"].([]interface{}); ok {
		urls := make([]string, len(v))
		for i, u := range v {
			urls[i], _ = u.(string)
		}
		listVal, diags := types.ListValueFrom(ctx, types.StringType, urls)
		resp.Diagnostics.Append(diags...)
		state.BlockListUrls = listVal
	} else {
		state.BlockListUrls = types.ListNull(types.StringType)
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

	if v, ok := response["maxLogFileDays"].(float64); ok {
		state.MaxLogFileDays = types.Int64Value(int64(v))
	} else {
		state.MaxLogFileDays = types.Int64Null()
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

	if v, ok := response["cacheMaximumEntries"].(float64); ok {
		state.CacheMaximumEntries = types.Int64Value(int64(v))
	} else {
		state.CacheMaximumEntries = types.Int64Null()
	}

	if v, ok := response["cacheMinimumRecordTtl"].(float64); ok {
		state.CacheMinimumRecordTtl = types.Int64Value(int64(v))
	} else {
		state.CacheMinimumRecordTtl = types.Int64Null()
	}

	if v, ok := response["cacheMaximumRecordTtl"].(float64); ok {
		state.CacheMaximumRecordTtl = types.Int64Value(int64(v))
	} else {
		state.CacheMaximumRecordTtl = types.Int64Null()
	}

	if v, ok := response["cacheNegativeRecordTtl"].(float64); ok {
		state.CacheNegativeRecordTtl = types.Int64Value(int64(v))
	} else {
		state.CacheNegativeRecordTtl = types.Int64Null()
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

	if v, ok := response["serveStale"].(bool); ok {
		state.ServeStale = types.BoolValue(v)
	} else {
		state.ServeStale = types.BoolNull()
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

	if v, ok := response["enableDnsOverUdpProxy"].(bool); ok {
		state.EnableDnsOverUdpProxy = types.BoolValue(v)
	} else {
		state.EnableDnsOverUdpProxy = types.BoolNull()
	}

	if v, ok := response["enableDnsOverTcpProxy"].(bool); ok {
		state.EnableDnsOverTcpProxy = types.BoolValue(v)
	} else {
		state.EnableDnsOverTcpProxy = types.BoolNull()
	}

	if v, ok := response["enableDnsOverHttp"].(bool); ok {
		state.EnableDnsOverHttp = types.BoolValue(v)
	} else {
		state.EnableDnsOverHttp = types.BoolNull()
	}

	if v, ok := response["enableDnsOverTls"].(bool); ok {
		state.EnableDnsOverTls = types.BoolValue(v)
	} else {
		state.EnableDnsOverTls = types.BoolNull()
	}

	if v, ok := response["enableDnsOverHttps"].(bool); ok {
		state.EnableDnsOverHttps = types.BoolValue(v)
	} else {
		state.EnableDnsOverHttps = types.BoolNull()
	}

	if v, ok := response["enableDnsOverHttp3"].(bool); ok {
		state.EnableDnsOverHttp3 = types.BoolValue(v)
	} else {
		state.EnableDnsOverHttp3 = types.BoolNull()
	}

	if v, ok := response["enableDnsOverQuic"].(bool); ok {
		state.EnableDnsOverQuic = types.BoolValue(v)
	} else {
		state.EnableDnsOverQuic = types.BoolNull()
	}

	if v, ok := response["dnsOverUdpProxyPort"].(float64); ok {
		state.DnsOverUdpProxyPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverUdpProxyPort = types.Int64Null()
	}

	if v, ok := response["dnsOverTcpProxyPort"].(float64); ok {
		state.DnsOverTcpProxyPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverTcpProxyPort = types.Int64Null()
	}

	if v, ok := response["dnsOverHttpPort"].(float64); ok {
		state.DnsOverHttpPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverHttpPort = types.Int64Null()
	}

	if v, ok := response["dnsOverTlsPort"].(float64); ok {
		state.DnsOverTlsPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverTlsPort = types.Int64Null()
	}

	if v, ok := response["dnsOverHttpsPort"].(float64); ok {
		state.DnsOverHttpsPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverHttpsPort = types.Int64Null()
	}

	if v, ok := response["dnsOverQuicPort"].(float64); ok {
		state.DnsOverQuicPort = types.Int64Value(int64(v))
	} else {
		state.DnsOverQuicPort = types.Int64Null()
	}

	if v, ok := response["dnsTlsCertificatePath"].(string); ok {
		state.DnsTlsCertificatePath = types.StringValue(v)
	} else {
		state.DnsTlsCertificatePath = types.StringNull()
	}

	if v, ok := response["dnsTlsCertificatePassword"].(string); ok {
		state.DnsTlsCertificatePassword = types.StringValue(v)
	} else {
		state.DnsTlsCertificatePassword = types.StringNull()
	}

	if v, ok := response["webServiceHttpPort"].(float64); ok {
		state.WebServiceHttpPort = types.Int64Value(int64(v))
	} else {
		state.WebServiceHttpPort = types.Int64Null()
	}

	if v, ok := response["webServiceTlsPort"].(float64); ok {
		state.WebServiceTlsPort = types.Int64Value(int64(v))
	} else {
		state.WebServiceTlsPort = types.Int64Null()
	}

	if v, ok := response["webServiceEnableTls"].(bool); ok {
		state.WebServiceEnableTls = types.BoolValue(v)
	} else {
		state.WebServiceEnableTls = types.BoolNull()
	}

	if v, ok := response["webServiceEnableHttp3"].(bool); ok {
		state.WebServiceEnableHttp3 = types.BoolValue(v)
	} else {
		state.WebServiceEnableHttp3 = types.BoolNull()
	}

	if v, ok := response["webServiceHttpToTlsRedirect"].(bool); ok {
		state.WebServiceHttpToTlsRedirect = types.BoolValue(v)
	} else {
		state.WebServiceHttpToTlsRedirect = types.BoolNull()
	}

	if v, ok := response["webServiceUseSelfSignedTlsCertificate"].(bool); ok {
		state.WebServiceUseSelfSignedTlsCertificate = types.BoolValue(v)
	} else {
		state.WebServiceUseSelfSignedTlsCertificate = types.BoolNull()
	}

	if v, ok := response["webServiceTlsCertificatePath"].(string); ok {
		state.WebServiceTlsCertificatePath = types.StringValue(v)
	} else {
		state.WebServiceTlsCertificatePath = types.StringNull()
	}

	if v, ok := response["webServiceTlsCertificatePassword"].(string); ok {
		state.WebServiceTlsCertificatePassword = types.StringValue(v)
	} else {
		state.WebServiceTlsCertificatePassword = types.StringNull()
	}

	if v, ok := response["webServiceRealIpHeader"].(string); ok {
		state.WebServiceRealIpHeader = types.StringValue(v)
	} else {
		state.WebServiceRealIpHeader = types.StringNull()
	}

	if v, ok := response["dnsOverHttpRealIpHeader"].(string); ok {
		state.DnsOverHttpRealIpHeader = types.StringValue(v)
	} else {
		state.DnsOverHttpRealIpHeader = types.StringNull()
	}

	if v, ok := response["proxyType"].(string); ok {
		state.ServerProxyType = types.StringValue(v)
	} else {
		state.ServerProxyType = types.StringNull()
	}

	if v, ok := response["proxyAddress"].(string); ok {
		state.ServerProxyAddress = types.StringValue(v)
	} else {
		state.ServerProxyAddress = types.StringNull()
	}

	if v, ok := response["proxyPort"].(float64); ok {
		state.ServerProxyPort = types.Int64Value(int64(v))
	} else {
		state.ServerProxyPort = types.Int64Null()
	}

	if v, ok := response["proxyUsername"].(string); ok {
		state.ServerProxyUsername = types.StringValue(v)
	} else {
		state.ServerProxyUsername = types.StringNull()
	}

	if v, ok := response["proxyPassword"].(string); ok {
		state.ServerProxyPassword = types.StringValue(v)
	} else {
		state.ServerProxyPassword = types.StringNull()
	}

	if v, ok := response["proxyBypass"].(string); ok {
		state.ServerProxyBypass = types.StringValue(v)
	} else {
		state.ServerProxyBypass = types.StringNull()
	}

	if v, ok := response["eDnsClientSubnet"].(bool); ok {
		state.EdnsClientSubnet = types.BoolValue(v)
	} else {
		state.EdnsClientSubnet = types.BoolNull()
	}

	if v, ok := response["eDnsClientSubnetIPv4PrefixLength"].(float64); ok {
		state.EdnsClientSubnetIpv4PrefixLength = types.Int64Value(int64(v))
	} else {
		state.EdnsClientSubnetIpv4PrefixLength = types.Int64Null()
	}

	if v, ok := response["eDnsClientSubnetIPv6PrefixLength"].(float64); ok {
		state.EdnsClientSubnetIpv6PrefixLength = types.Int64Value(int64(v))
	} else {
		state.EdnsClientSubnetIpv6PrefixLength = types.Int64Null()
	}

	if v, ok := response["eDnsClientSubnetIpv4Override"].(string); ok {
		state.EdnsClientSubnetIpv4Override = types.StringValue(v)
	} else {
		state.EdnsClientSubnetIpv4Override = types.StringNull()
	}

	if v, ok := response["eDnsClientSubnetIpv6Override"].(string); ok {
		state.EdnsClientSubnetIpv6Override = types.StringValue(v)
	} else {
		state.EdnsClientSubnetIpv6Override = types.StringNull()
	}

	if v, ok := response["defaultResponsiblePerson"].(string); ok {
		state.DefaultResponsiblePerson = types.StringValue(v)
	} else {
		state.DefaultResponsiblePerson = types.StringNull()
	}

	if v, ok := response["useSoaSerialDateScheme"].(bool); ok {
		state.UseSoaSerialDateScheme = types.BoolValue(v)
	} else {
		state.UseSoaSerialDateScheme = types.BoolNull()
	}

	if v, ok := response["dnsAppsEnableAutomaticUpdate"].(bool); ok {
		state.DnsAppsEnableAutomaticUpdate = types.BoolValue(v)
	} else {
		state.DnsAppsEnableAutomaticUpdate = types.BoolNull()
	}

	if v, ok := response["enableUdpSocketPool"].(bool); ok {
		state.EnableUdpSocketPool = types.BoolValue(v)
	} else {
		state.EnableUdpSocketPool = types.BoolNull()
	}

	if v, ok := response["quicIdleTimeout"].(float64); ok {
		state.QuicIdleTimeout = types.Int64Value(int64(v))
	} else {
		state.QuicIdleTimeout = types.Int64Null()
	}

	if v, ok := response["quicMaxInboundStreams"].(float64); ok {
		state.QuicMaxInboundStreams = types.Int64Value(int64(v))
	} else {
		state.QuicMaxInboundStreams = types.Int64Null()
	}

	if v, ok := response["loggingType"].(string); ok {
		state.LoggingType = types.StringValue(v)
	} else {
		state.LoggingType = types.StringNull()
	}

	if v, ok := response["ignoreResolverLogs"].(bool); ok {
		state.IgnoreResolverLogs = types.BoolValue(v)
	} else {
		state.IgnoreResolverLogs = types.BoolNull()
	}

	if v, ok := response["useLocalTime"].(bool); ok {
		state.UseLocalTime = types.BoolValue(v)
	} else {
		state.UseLocalTime = types.BoolNull()
	}

	if v, ok := response["logFolder"].(string); ok {
		state.LogFolder = types.StringValue(v)
	} else {
		state.LogFolder = types.StringNull()
	}

	if v, ok := response["enableInMemoryStats"].(bool); ok {
		state.EnableInMemoryStats = types.BoolValue(v)
	} else {
		state.EnableInMemoryStats = types.BoolNull()
	}

	if v, ok := response["maxStatFileDays"].(float64); ok {
		state.MaxStatFileDays = types.Int64Value(int64(v))
	} else {
		state.MaxStatFileDays = types.Int64Null()
	}

	if v, ok := response["maxConcurrentResolutionsPerCore"].(float64); ok {
		state.MaxConcurrentResolutionsPerCore = types.Int64Value(int64(v))
	} else {
		state.MaxConcurrentResolutionsPerCore = types.Int64Null()
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
