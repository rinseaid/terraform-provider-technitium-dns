package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium-dns/internal/client"
)

var _ datasource.DataSource = &dnsAppsDataSource{}

func NewDNSAppsDataSource() datasource.DataSource {
	return &dnsAppsDataSource{}
}

type dnsAppsDataSource struct {
	client *client.Client
}

type dnsAppsDataSourceModel struct {
	Apps []dnsAppModel `tfsdk:"apps"`
}

type dnsAppModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

func (d *dnsAppsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_apps"
}

func (d *dnsAppsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists installed DNS apps on a Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"apps": schema.ListNestedAttribute{
				Description: "List of installed DNS apps.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the DNS app.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "The installed version of the DNS app.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dnsAppsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsAppsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DNS Apps",
			fmt.Sprintf("Could not list DNS apps: %s", err),
		)
		return
	}

	var state dnsAppsDataSourceModel

	appList, ok := result["apps"].([]interface{})
	if !ok {
		state.Apps = []dnsAppModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	for _, entry := range appList {
		app, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		model := dnsAppModel{}
		if v, ok := app["name"].(string); ok {
			model.Name = types.StringValue(v)
		}
		if v, ok := app["version"].(string); ok {
			model.Version = types.StringValue(v)
		}
		state.Apps = append(state.Apps, model)
	}

	if state.Apps == nil {
		state.Apps = []dnsAppModel{}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
