package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &dnsRecordsDataSource{}

func NewDNSRecordsDataSource() datasource.DataSource {
	return &dnsRecordsDataSource{}
}

type dnsRecordsDataSource struct {
	client *client.Client
}

type dnsRecordsDataSourceModel struct {
	Zone    types.String          `tfsdk:"zone"`
	Domain  types.String          `tfsdk:"domain"`
	Type    types.String          `tfsdk:"type"`
	Records []dnsRecordEntryModel `tfsdk:"records"`
}

type dnsRecordEntryModel struct {
	Domain   types.String `tfsdk:"domain"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Comments types.String `tfsdk:"comments"`
	Priority types.Int64  `tfsdk:"priority"`
	Weight   types.Int64  `tfsdk:"weight"`
	Port     types.Int64  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
}

func (d *dnsRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *dnsRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads DNS records from a Technitium DNS Server authoritative zone.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Description: "The authoritative zone name.",
				Required:    true,
			},
			"domain": schema.StringAttribute{
				Description: "Filter records by domain name. When omitted, all records in the zone are returned.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Filter records by DNS record type (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA, FWD).",
				Optional:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "List of matching DNS records.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The fully qualified domain name of the record.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The DNS record type.",
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "The record value.",
							Computed:    true,
						},
						"ttl": schema.Int64Attribute{
							Description: "Time to live in seconds.",
							Computed:    true,
						},
						"disabled": schema.BoolAttribute{
							Description: "Whether the record is disabled.",
							Computed:    true,
						},
						"comments": schema.StringAttribute{
							Description: "Comments for the record.",
							Computed:    true,
						},
						"priority": schema.Int64Attribute{
							Description: "Priority value for MX and SRV records.",
							Computed:    true,
						},
						"weight": schema.Int64Attribute{
							Description: "Weight value for SRV records.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Port number for SRV records.",
							Computed:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "Forwarding protocol for FWD records.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dnsRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnsRecordsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := config.Zone.ValueString()

	// When domain is not specified, query the zone apex with listZone=true
	// to get all records. When domain is specified, query that specific domain.
	domain := zone
	listZone := true
	if !config.Domain.IsNull() && !config.Domain.IsUnknown() {
		domain = config.Domain.ValueString()
		listZone = false
	}

	typeFilter := ""
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		typeFilter = config.Type.ValueString()
	}

	tflog.Debug(ctx, "Reading DNS records", map[string]interface{}{
		"zone":     zone,
		"domain":   domain,
		"type":     typeFilter,
		"listZone": listZone,
	})

	response, err := d.client.GetRecords(ctx, domain, zone, listZone)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DNS Records",
			fmt.Sprintf("Could not read records for zone %s: %s", zone, err.Error()),
		)
		return
	}

	recordsList, ok := response["records"].([]interface{})
	if !ok {
		// No records found, return empty list.
		config.Records = []dnsRecordEntryModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	var entries []dnsRecordEntryModel
	for _, item := range recordsList {
		rec, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		recType := stringFromMap(rec, "type")

		// Skip DNSSEC-internal record types.
		switch recType {
		case "RRSIG", "NSEC", "NSEC3", "NSEC3PARAM", "DNSKEY":
			continue
		}

		// Apply type filter.
		if typeFilter != "" && !strings.EqualFold(recType, typeFilter) {
			continue
		}

		value := recordValueFromRData(rec, recType)

		entry := dnsRecordEntryModel{
			Domain:   types.StringValue(stringFromMap(rec, "name")),
			Type:     types.StringValue(recType),
			Value:    types.StringValue(value),
			Disabled: types.BoolValue(boolFromMap(rec, "disabled")),
		}

		if ttl, ok := rec["ttl"].(float64); ok {
			entry.TTL = types.Int64Value(int64(ttl))
		} else {
			entry.TTL = types.Int64Value(0)
		}

		if comments := stringFromMap(rec, "comments"); comments != "" {
			entry.Comments = types.StringValue(comments)
		} else {
			entry.Comments = types.StringNull()
		}

		rData, _ := rec["rData"].(map[string]interface{})
		entry.Priority = types.Int64Null()
		entry.Weight = types.Int64Null()
		entry.Port = types.Int64Null()
		entry.Protocol = types.StringNull()

		if rData != nil {
			switch recType {
			case "MX":
				if pref, ok := rData["preference"].(float64); ok && pref > 0 {
					entry.Priority = types.Int64Value(int64(pref))
				}
			case "SRV":
				if prio, ok := rData["priority"].(float64); ok {
					entry.Priority = types.Int64Value(int64(prio))
				}
				if w, ok := rData["weight"].(float64); ok {
					entry.Weight = types.Int64Value(int64(w))
				}
				if p, ok := rData["port"].(float64); ok {
					entry.Port = types.Int64Value(int64(p))
				}
			case "FWD":
				if proto, ok := rData["protocol"].(string); ok && proto != "" {
					entry.Protocol = types.StringValue(proto)
				}
			}
		}

		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []dnsRecordEntryModel{}
	}

	config.Records = entries
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
