// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LicenseResource{}
var _ resource.ResourceWithImportState = &LicenseResource{}

func NewLicenseResource() resource.Resource {
	return &LicenseResource{}
}

// LicenseResource defines the resource implementation.
type LicenseResource struct {
	client   Client
	endpoint string
	token    string
}

// LicenseResourceModel describes the resource data model.
type LicenseResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Product         types.String `tfsdk:"product"`
	Type            types.String `tfsdk:"type"`
	Expiration      types.String `tfsdk:"expiration"`
	Status          types.String `tfsdk:"status"`
	MaxInstances    types.Int64  `tfsdk:"max_instances"`
	CreatorID       types.String `tfsdk:"creator_id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Offline         types.Bool   `tfsdk:"offline"`
	Created         types.String `tfsdk:"created"`
	MetaLastUpdated types.String `tfsdk:"meta_last_updated"`
	MetaCreatedAt   types.String `tfsdk:"meta_created_at"`
	MetaVersionID   types.String `tfsdk:"meta_version_id"`
	Issuer          types.String `tfsdk:"issuer"`
	InfoHosting     types.String `tfsdk:"info_hosting"`
	JWT             types.String `tfsdk:"jwt"`
}

func (r *LicenseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (r *LicenseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Aidbox license",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"product": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"expiration": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"max_instances": schema.Int64Attribute{
				Computed: true,
			},
			"creator_id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"offline": schema.BoolAttribute{
				Computed: true,
			},
			"created": schema.StringAttribute{
				Computed: true,
			},
			"meta_last_updated": schema.StringAttribute{
				Computed: true,
			},
			"meta_created_at": schema.StringAttribute{
				Computed: true,
			},
			"meta_version_id": schema.StringAttribute{
				Computed: true,
			},
			"issuer": schema.StringAttribute{
				Computed: true,
			},
			"info_hosting": schema.StringAttribute{
				Computed: true,
			},
			"jwt": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *LicenseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = data.Client
	r.endpoint = data.Endpoint
	r.token = data.Token
}

func (r *LicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LicenseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	yamlRequestBody, err := createYAMLRequestBody(data, r.token)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create YAML request body", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("API Request %s", yamlRequestBody))

	httpReq, err := http.NewRequest("POST", r.endpoint, strings.NewReader(yamlRequestBody))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", err.Error())
		return
	}
	httpReq.Header.Set("Content-Type", "text/yaml")
	httpReq.Header.Set("Accept", "text/yaml")

	apiResp, err := r.client.CreateLicense(ctx, data.Name.ValueString(), data.Product.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API call failed", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("API Response %s", apiResp.JWT))
	data.ID = basetypes.NewStringValue(apiResp.License.ID)
	data.Name = basetypes.NewStringValue(apiResp.License.Name)
	data.Product = basetypes.NewStringValue(apiResp.License.Product)
	data.Type = basetypes.NewStringValue(apiResp.License.Type)
	data.Expiration = basetypes.NewStringValue(apiResp.License.Expiration)
	data.Status = basetypes.NewStringValue(apiResp.License.Status)
	data.MaxInstances = basetypes.NewInt64Value(int64(apiResp.License.MaxInstances))
	data.CreatorID = basetypes.NewStringValue(apiResp.License.Creator.ID)
	data.ProjectID = basetypes.NewStringValue(apiResp.License.Project.ID)
	data.Offline = basetypes.NewBoolValue(apiResp.License.Offline)
	data.Created = basetypes.NewStringValue(apiResp.License.Created)
	data.MetaLastUpdated = basetypes.NewStringValue(apiResp.License.Meta.LastUpdated)
	data.MetaCreatedAt = basetypes.NewStringValue(apiResp.License.Meta.CreatedAt)
	data.MetaVersionID = basetypes.NewStringValue(apiResp.License.Meta.VersionID)
	data.Issuer = basetypes.NewStringValue(apiResp.License.Issuer)
	data.InfoHosting = basetypes.NewStringValue(apiResp.License.Info.Hosting)
	data.JWT = basetypes.NewStringValue(apiResp.JWT)

	// Process data further or set it in the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LicenseResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LicenseResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LicenseResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *LicenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createYAMLRequestBody(data LicenseResourceModel, token string) (string, error) {
	requestBody := map[string]interface{}{
		"method": "portal.portal/issue-license",
		"params": map[string]string{
			"token":   token,
			"name":    data.Name.ValueString(),
			"product": data.Product.ValueString(),
			"type":    data.Type.ValueString(),
		},
	}

	yamlData, err := yaml.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

// Assuming your response structure matches this model
type APIResponse struct {
	Result struct {
		Cluster    interface{} `yaml:"cluster"`
		Deployment interface{} `yaml:"deployment"`
		License    struct {
			Offline bool `yaml:"offline"`
			Meta    struct {
				LastUpdated string `yaml:"lastUpdated"`
				CreatedAt   string `yaml:"createdAt"`
				VersionID   string `yaml:"versionId"`
			} `yaml:"meta"`
			Creator struct {
				ID           string `yaml:"id"`
				ResourceType string `yaml:"resourceType"`
			} `yaml:"creator"`
			Name         string `yaml:"name"`
			Expiration   string `yaml:"expiration"`
			Type         string `yaml:"type"`
			Created      string `yaml:"created"`
			ResourceType string `yaml:"resourceType"`
			MaxInstances int    `yaml:"max-instances"`
			Product      string `yaml:"product"`
			Project      struct {
				ID           string `yaml:"id"`
				ResourceType string `yaml:"resourceType"`
			} `yaml:"project"`
			Status string `yaml:"status"`
			ID     string `yaml:"id"`
			Info   struct {
				Hosting string `yaml:"hosting"`
			} `yaml:"info"`
			Issuer     string `yaml:"issuer"`
			Additional struct {
				ExpirationDays int    `yaml:"expiration-days"`
				BoxURL         string `yaml:"box-url"`
			} `yaml:"additional"`
		} `yaml:"license"`
		JWT string `yaml:"jwt"`
	} `yaml:"result"`
}

// Example function to parse YAML
func parseYAMLResponse(body []byte) (*APIResponse, error) {
	var resp APIResponse
	if err := yaml.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
