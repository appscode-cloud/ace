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

func (c *Client) ImportCluster(basicInfo ClusterBasicInfo, opts ProviderOptions, installFluxCD bool) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath, err := getProviderSpecificAPIPath(org, opts, "import")
	if err != nil {
		return nil, err
	}
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "credential", value: opts.Credential},
		{key: "install-fluxcd", value: fmt.Sprintf("%v", installFluxCD)},
	})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(basicInfo)
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

func (c *Client) ListClusters() (*v1alpha1.ClusterInfoList, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s", org)

	var clusters v1alpha1.ClusterInfoList
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &clusters)
	if err != nil {
		return nil, err
	}
	return &clusters, nil
}

func (c *Client) GetCluster(clusterName string) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/status", org, clusterName)

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) ConnectCluster(clusterName, credential string) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/connect", org, clusterName)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "credential", value: credential},
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

func (c *Client) ReconfigureCluster(clusterName string, installFluxCD bool) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/reconfigure", org, clusterName)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "install-fluxcd", value: fmt.Sprintf("%v", installFluxCD)},
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

func (c *Client) RemoveCluster(clusterName string, removeFluxCD bool) error {
	org, err := c.getOrganization()
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/remove", org, clusterName)
	apiPath, err = setQueryParams(apiPath, []queryParams{
		{key: "remove-fluxcd", value: fmt.Sprintf("%v", removeFluxCD)},
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
