package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	clustermodel "go.bytebuilders.dev/resource-model/apis/cluster"
	"go.bytebuilders.dev/resource-model/apis/cluster/v1alpha1"
)

func (c *Client) CheckClusterExistence(opts clustermodel.ProviderOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath, err := getProviderSpecificAPIPath(org, opts, "check")
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider options. err: %w", err)
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, bytes.NewReader(body), &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) ImportCluster(opts clustermodel.ImportOptions, responseID string) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath, err := getProviderSpecificAPIPath(org, opts.Provider, "import")
	if err != nil {
		return nil, err
	}
	params := make([]queryParams, 0)
	if responseID != "" {
		params = append(params, queryParams{key: "response-id", value: responseID})
	}
	apiPath, err = setQueryParams(apiPath, params)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(opts)
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

func (c *Client) ListClusters(opts clustermodel.ListOptions) (*v1alpha1.ClusterInfoList, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s", org)

	params := make([]queryParams, 0)
	if opts.Provider != "" {
		params = append(params, queryParams{key: "provider", value: opts.Provider})
	}

	apiPath, err = setQueryParams(apiPath, params)
	if err != nil {
		return nil, err
	}
	var clusters v1alpha1.ClusterInfoList
	err = c.getParsedResponse(http.MethodGet, apiPath, jsonHeader, nil, &clusters)
	if err != nil {
		return nil, err
	}
	return &clusters, nil
}

func (c *Client) GetCluster(opts clustermodel.GetOptions) (*v1alpha1.ClusterInfo, error) {
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

func (c *Client) ConnectCluster(opts clustermodel.ConnectOptions) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/connect", org, opts.Name)

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal connect options. err: %w", err)
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, bytes.NewReader(body), &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) ReconfigureCluster(opts clustermodel.ReconfigureOptions, responseID string) (*v1alpha1.ClusterInfo, error) {
	org, err := c.getOrganization()
	if err != nil {
		return nil, err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/reconfigure", org, opts.Name)

	params := make([]queryParams, 0)
	if responseID != "" {
		params = append(params, queryParams{key: "response-id", value: responseID})
	}
	apiPath, err = setQueryParams(apiPath, params)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal reconfigure options. err: %w", err)
	}

	var cluster v1alpha1.ClusterInfo
	err = c.getParsedResponse(http.MethodPost, apiPath, jsonHeader, bytes.NewReader(body), &cluster)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) RemoveCluster(opts clustermodel.RemovalOptions, responseID string) error {
	org, err := c.getOrganization()
	if err != nil {
		return err
	}
	apiPath := fmt.Sprintf("/clustersv2/%s/%s/remove", org, opts.Name)

	params := make([]queryParams, 0)
	if responseID != "" {
		params = append(params, queryParams{key: "response-id", value: responseID})
	}
	apiPath, err = setQueryParams(apiPath, params)
	if err != nil {
		return err
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to unmarshal remove options. err: %w", err)
	}
	_, err = c.getResponse(http.MethodPost, apiPath, jsonHeader, bytes.NewReader(body))
	if err != nil {
		return err
	}
	return nil
}

func getProviderSpecificAPIPath(org string, opts clustermodel.ProviderOptions, suffix string) (string, error) {
	var apiPath string
	switch lower(opts.Name) {
	case lower(string(v1alpha1.ProviderLinode)):
		if opts.ClusterID == "" {
			return "", fmt.Errorf("missing linode cluster ID")
		}
		apiPath = fmt.Sprintf("/clouds2/%s/providers/linode/%s", org, opts.ClusterID)
	case lower(string(v1alpha1.ProviderGeneric)):
		apiPath = fmt.Sprintf("/clouds2/%s/providers/generic", org)
	default:
		return "", fmt.Errorf("import is not supported for provider %s", opts.Name)
	}
	return filepath.Join(apiPath, suffix), nil
}

func lower(s string) string {
	return strings.ToLower(s)
}
