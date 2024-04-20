// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"net/http"
	"os" // Import for environment variables

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &ScaffoldingProvider{}
var _ provider.ProviderWithFunctions = &ScaffoldingProvider{}

type ScaffoldingProvider struct {
	version string
}

type ScaffoldingProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *ScaffoldingProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aidbox"
	resp.Version = p.version
}

func (p *ScaffoldingProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Aidbox RPC API endpoint",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Aidbox token",
				Optional:            true,
			},
		},
	}
}

func (p *ScaffoldingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default endpoint if not provided
	if data.Endpoint.IsNull() || data.Endpoint.IsUnknown() || data.Endpoint.ValueString() == "" {
		defaultEndpoint := basetypes.NewStringValue("https://aidbox.app/rpc")
		data.Endpoint = defaultEndpoint
	}

	// Handle token; get from environment variable if not provided
	if data.Token.IsNull() || data.Token.IsUnknown() || data.Token.ValueString() == "" {
		tokenEnv := os.Getenv("AIDBOX_TOKEN")
		if tokenEnv != "" {
			data.Token = basetypes.NewStringValue(tokenEnv)
		} else {
			resp.Diagnostics.AddError(
				"No Token Provided",
				"Please provide a 'token' in the provider configuration or through the 'AIDBOX_TOKEN' environment variable.",
			)
			return
		}
	}

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ScaffoldingProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *ScaffoldingProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *ScaffoldingProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaffoldingProvider{
			version: version,
		}
	}
}
