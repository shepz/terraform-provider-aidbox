package aidboxclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"
)

type AidboxHTTPClient struct {
	Endpoint string
	Token    string
	Client   *http.Client
}

type Creator struct {
	ID           string `yaml:"id"`
	ResourceType string `yaml:"resourceType"`
}

type Project struct {
	ID           string `yaml:"id"`
	ResourceType string `yaml:"resourceType"`
}

type Info struct {
	Hosting string `yaml:"hosting"`
}

type Meta struct {
	LastUpdated string `yaml:"lastUpdated"`
	CreatedAt   string `yaml:"createdAt"`
	VersionID   string `yaml:"versionId"`
}

type Additional struct {
	ExpirationDays int     `yaml:"expiration-days"`
	BoxURL         *string `yaml:"box-url"`
}

type License struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	Product      string     `yaml:"product"`
	Type         string     `yaml:"type"`
	Expiration   string     `yaml:"expiration"`
	Status       string     `yaml:"status"`
	MaxInstances int        `yaml:"max-instances"`
	Creator      Creator    `yaml:"creator"`
	Project      Project    `yaml:"project"`
	Offline      bool       `yaml:"offline"`
	Created      string     `yaml:"created"`
	Meta         Meta       `yaml:"meta"`
	Issuer       string     `yaml:"issuer"`
	Info         Info       `yaml:"info"`
	Additional   Additional `yaml:"additional"`
}

// LicenseResponse includes the License and JWT token.
type LicenseResponse struct {
	License License
	JWT     string
}

// APIResponse maps the YAML response from the Aidbox API.
type APIResponse struct {
	Result struct {
		License License `yaml:"license"`
		JWT     string  `yaml:"jwt"`
	}
}

func NewClient(endpoint, token string) *AidboxHTTPClient {
	return &AidboxHTTPClient{
		Endpoint: endpoint,
		Token:    token,
		Client:   http.DefaultClient,
	}
}

func (c *AidboxHTTPClient) CreateLicense(ctx context.Context, name, product, licenseType string) (LicenseResponse, error) {
	tflog.Debug(ctx, "Creating license", map[string]interface{}{
		"name":    name,
		"product": product,
		"type":    licenseType,
	})

	requestBody := map[string]interface{}{
		"method": "portal.portal/issue-license",
		"params": map[string]interface{}{
			"token":   c.Token,
			"name":    name,
			"product": product,
			"type":    licenseType,
		},
	}

	yamlData, err := yaml.Marshal(requestBody)
	if err != nil {
		tflog.Error(ctx, "Failed to create YAML request body", map[string]interface{}{"error": err})
		return LicenseResponse{}, fmt.Errorf("failed to create YAML request body: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, strings.NewReader(string(yamlData)))
	if err != nil {
		tflog.Error(ctx, "Failed to create HTTP request", map[string]interface{}{"error": err})
		return LicenseResponse{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "text/yaml")
	req.Header.Set("Accept", "text/yaml")

	resp, err := c.Client.Do(req)
	if err != nil {
		tflog.Error(ctx, "API call failed", map[string]interface{}{"error": err})
		return LicenseResponse{}, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		tflog.Error(ctx, "Failed to read response body", map[string]interface{}{"error": err})
		return LicenseResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		tflog.Error(ctx, "API response error", map[string]interface{}{
			"status": resp.Status,
			"body":   string(bodyBytes),
		})
		return LicenseResponse{}, fmt.Errorf("API response error: %s; Body: %s", resp.Status, string(bodyBytes))
	}

	var apiResp APIResponse
	if err := yaml.Unmarshal(bodyBytes, &apiResp); err != nil {
		tflog.Error(ctx, "Failed to parse YAML response", map[string]interface{}{"error": err, "body": string(bodyBytes)})
		return LicenseResponse{}, fmt.Errorf("failed to parse YAML response: %s; Body: %s", err, string(bodyBytes))
	}

	return LicenseResponse{
		License: apiResp.Result.License,
		JWT:     apiResp.Result.JWT,
	}, nil
}
