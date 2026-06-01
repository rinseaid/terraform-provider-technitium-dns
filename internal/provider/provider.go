package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rinseaid/terraform-provider-technitium/internal/client"
)

var _ provider.Provider = &TechnitiumProvider{}

type TechnitiumProvider struct {
	version string
}

type TechnitiumProviderModel struct {
	ServerURL types.String `tfsdk:"server_url"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	APIToken  types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TechnitiumProvider{
			version: version,
		}
	}
}

func (p *TechnitiumProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "technitium"
	resp.Version = p.version
}

func (p *TechnitiumProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Technitium DNS Server resources.",
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				Description: "URL of the Technitium DNS Server API (e.g. http://localhost:5380). " +
					"Can also be set with the TECHNITIUM_SERVER_URL environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Technitium DNS Server authentication. " +
					"Can also be set with the TECHNITIUM_USERNAME environment variable. " +
					"Required if api_token is not set.",
				Optional:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Technitium DNS Server authentication. " +
					"Can also be set with the TECHNITIUM_PASSWORD environment variable. " +
					"Required if api_token is not set.",
				Optional:  true,
				Sensitive: true,
			},
			"api_token": schema.StringAttribute{
				Description: "API token for Technitium DNS Server authentication. " +
					"Can also be set with the TECHNITIUM_API_TOKEN environment variable. " +
					"Alternative to username/password authentication.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *TechnitiumProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TechnitiumProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve server_url: config value takes precedence over env var.
	serverURL := config.ServerURL.ValueString()
	if config.ServerURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Unknown Technitium Server URL",
			"The provider cannot create the API client because the server_url value is unknown. "+
				"Set the value statically in the configuration or use the TECHNITIUM_SERVER_URL environment variable.",
		)
	}
	if serverURL == "" {
		serverURL = os.Getenv("TECHNITIUM_SERVER_URL")
	}
	if serverURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Missing Technitium Server URL",
			"The provider requires a server URL to connect to the Technitium DNS Server API. "+
				"Set the server_url attribute or the TECHNITIUM_SERVER_URL environment variable.",
		)
	}

	// Resolve credentials: config values take precedence over env vars.
	username := config.Username.ValueString()
	if username == "" {
		username = os.Getenv("TECHNITIUM_USERNAME")
	}

	password := config.Password.ValueString()
	if password == "" {
		password = os.Getenv("TECHNITIUM_PASSWORD")
	}

	apiToken := config.APIToken.ValueString()
	if apiToken == "" {
		apiToken = os.Getenv("TECHNITIUM_API_TOKEN")
	}

	// Validate that at least one auth method is provided.
	if apiToken == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing Technitium Credentials",
			"The provider requires either an api_token or both username and password to authenticate. "+
				"Set the appropriate attributes or environment variables: "+
				"TECHNITIUM_API_TOKEN, or TECHNITIUM_USERNAME and TECHNITIUM_PASSWORD.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the API client.
	var c *client.Client
	var err error

	if apiToken != "" {
		c, err = client.NewWithToken(serverURL, apiToken)
	} else {
		c, err = client.NewWithCredentials(serverURL, username, password)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Technitium API Client",
			"An unexpected error occurred when creating the Technitium API client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *TechnitiumProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSZoneResource,
		NewDNSRecordResource,
		NewDHCPScopeResource,
		NewDHCPReservedLeaseResource,
		NewAllowedZoneResource,
		NewBlockedZoneResource,
	}
}

func (p *TechnitiumProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDNSZonesDataSource,
		NewDNSRecordsDataSource,
		NewDHCPScopesDataSource,
		NewDHCPLeasesDataSource,
		NewAllowedZonesDataSource,
		NewBlockedZonesDataSource,
	}
}
