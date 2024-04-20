// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &LicenseResource{}
var _ resource.ResourceWithImportState = &LicenseResource{}

func NewLicenseResource() resource.Resource {
	return &LicenseResource{}
}

// LicenseResource defines the resource implementation.
type LicenseResource struct {
	client   *http.Client
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
				Computed: true,
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

	tflog.Debug(ctx, "Creating new license", map[string]interface{}{"name": data.Name.String()})

	// Manually set default values if not provided
	if data.Product.IsNull() || data.Product.IsUnknown() {
		data.Product = basetypes.NewStringValue("aidbox")
	}

	// Prepare the API request body
	requestBody := fmt.Sprintf("method: portal.portal/issue-license\nparams:\n  token: %s\n  name: %s\n  product: %s\n  type: %s",
		r.token, data.Name.String(), data.Product.String(), data.Type.String())

	tflog.Debug(ctx, "API Request Body", map[string]interface{}{"request": requestBody})

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/rpc", r.endpoint), strings.NewReader(requestBody))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create HTTP request", fmt.Sprintf("Error: %s", err))
		return
	}

	httpReq.Header.Add("Content-Type", "text/yaml")
	httpReq.Header.Add("Accept", "text/yaml")

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("API Call Failed", fmt.Sprintf("Could not issue a license: %s", err))
		return
	}
	defer httpResp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read response body", fmt.Sprintf("Error: %s", err))
		return
	}

	// Check if response body is empty
	if len(bodyBytes) == 0 {
		resp.Diagnostics.AddError("Empty Response", "The API response is empty, expected JSON.")
		return
	}

	// Log the response body
	tflog.Debug(ctx, "API Response Body", map[string]interface{}{"response": string(bodyBytes)})

	// Parse JSON
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		resp.Diagnostics.AddError("Failed to parse response body", fmt.Sprintf("Error parsing JSON: %s; Response Body: %s", err, string(bodyBytes)))
		return
	}

	// Set the ID and other computed attributes from the API response
	if data.ID.String() == "" {
		// Assuming that the API response contains an 'id' for the license
		data.ID = basetypes.NewStringValue("extracted-id-from-response") // Update this line to extract the actual ID
	}

	// Log the creation
	tflog.Debug(ctx, "Created new license", map[string]interface{}{"license_id": data.ID.String()})

	// Save data into Terraform state
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
			"name":    data.Name.String(),
			"product": data.Product.String(),
			"type":    data.Type.String(),
		},
	}

	yamlData, err := yaml.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}
