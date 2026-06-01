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
	DnsServerDomain  types.String `tfsdk:"dns_server_domain"`
	DnssecValidation types.Bool   `tfsdk:"dnssec_validation"`
	Recursion        types.String `tfsdk:"recursion"`
	PreferIPv6       types.Bool   `tfsdk:"prefer_ipv6"`
	EnableBlocking   types.Bool   `tfsdk:"enable_blocking"`
	BlockingType     types.String `tfsdk:"blocking_type"`
	Forwarders       types.List   `tfsdk:"forwarders"`
	ForwarderProtocol types.String `tfsdk:"forwarder_protocol"`
	EnableLogging    types.Bool   `tfsdk:"enable_logging"`
	LogQueries       types.Bool   `tfsdk:"log_queries"`
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
