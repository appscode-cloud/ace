package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"
)

type ClusterBasicInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type ProviderOptions struct {
	Credential string
	Provider   string
	ClusterID  string
}

type ComponentOptions struct {
	FluxCD        bool
	LicenseServer bool
}

func (c *Client) CheckClusterExistence(opts ProviderOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath, err := getProviderSpecificAPIPath(org, opts, "check")
	if err != nil {
		return nil, err
	}
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "credential", value: opts.Credential},
	})
	if err != nil {
		return nil, err
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

type ClusterImportOptions struct {
	BasicInfo  ClusterBasicInfo
	Provider   ProviderOptions
	Components ComponentOptions
}

func (c *Client) ImportCluster(opts ClusterImportOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath, err := getProviderSpecificAPIPath(org, opts.Provider, "import")
	if err != nil {
		return nil, err
	}
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "credential", value: opts.Provider.Credential},
		{key: "install-fluxcd", value: fmt.Sprintf("%v", opts.Components.FluxCD)},
		{key: "install-license-server", value: fmt.Sprintf("%v", opts.Components.LicenseServer)},
	})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(opts.BasicInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cluster basic info. err: %w", err)
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, bytes.NewReader(body), &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

type ClusterListOptions struct {
	Provider string
}

func (c *Client) ListClusters(listOptions *ClusterListOptions) (*v1alpha1.ClusterInfoList, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s", org)

	if listOptions != nil {
		apiPath, err = setQueryParams(apiPath, listOptions.formatAsQueryParams())
		if err != nil {
			return nil, err
		}
	}

	var clusters v1alpha1.ClusterInfoList
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &clusters)
	if err != nil {
		return nil, err
	}
	return &clusters, nil
}

func (opts *ClusterListOptions) formatAsQueryParams() []queryParams {
	var params []queryParams
	if opts.Provider != "" {
		params = append(params, queryParams{key: "provider", value: opts.Provider})
	}
	return params
}

type ClusterGetOptions struct {
	Name string
}

func (c *Client) GetCluster(opts ClusterGetOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/status", org, opts.Name)

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

type ClusterConnectOptions struct {
	Name       string
	Credential string
}

func (c *Client) ConnectCluster(opts ClusterConnectOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/connect", org, opts.Name)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "credential", value: opts.Credential},
	})
	if err != nil {
		return nil, err
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, nil, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

type ClusterReconfigureOptions struct {
	Name       string
	Components ComponentOptions
}

func (c *Client) ReconfigureCluster(opts ClusterReconfigureOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/reconfigure", org, opts.Name)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "install-fluxcd", value: fmt.Sprintf("%v", opts.Components.FluxCD)},
		{key: "install-license-server", value: fmt.Sprintf("%v", opts.Components.LicenseServer)},
	})
	if err != nil {
		return nil, err
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, nil, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

type ClusterRemovalOptions struct {
	Name       string
	Components ComponentOptions
}

func (c *Client) RemoveCluster(opts ClusterRemovalOptions) error {
	org, err := c.getOrganization()
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/remove", org, opts.Name)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "remove-fluxcd", value: fmt.Sprintf("%v", opts.Components.FluxCD)},
		{key: "remove-license-server", value: fmt.Sprintf("%v", opts.Components.LicenseServer)},
	})
	if err != nil {
		return err
	}

	_, err = c.getResponse(http.MethodPost, apiPath, jsonHeader, nil)
	if err != nil {
		return err
	}
	return nil
}

func getProviderSpecificAPIPath(org string, opts ProviderOptions, suffix string) (string, error) {
	var apiPath string
	switch lower(opts.Provider) {
	case lower(string(v1alpha1.ProviderLinode)):
		if opts.ClusterID == "" {
			return "", fmt.Errorf("missing linode cluster ID")
		}
		apiPath = fmt.Sprintf("/clouds2/%s/providers/linode/%s", org, opts.ClusterID)
	default:
		return "", fmt.Errorf("import is not supported for provider %s", opts.Provider)
	}
	return filepath.Join(apiPath, suffix), nil
}

func lower(s string) string {
	return strings.ToLower(s)
}
