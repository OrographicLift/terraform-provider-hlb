// File: hlb/load_balancer.go

package hlb

import (
	"encoding/json"
	"fmt"
)

type LoadBalancer struct {
	ID                           string            `json:"id"`
    AccountID                    string            `json:"accountId"`
	URI                          string            `json:"uri"`
	Name                         string            `json:"name"`
	Internal                     bool              `json:"internal"`
	State                        string            `json:"state"`
	DNSName                      string            `json:"dnsName"`
	ZoneID                       string            `json:"zoneId"`
	CreatedAt                    string            `json:"createdAt"`
	UpdatedAt                    string            `json:"updatedAt"`
	Subnets                      []string          `json:"subnets"`
	SecurityGroups               []string          `json:"securityGroups"`
	AccessLogs                   *AccessLogs       `json:"accessLogs,omitempty"`
	EnableDeletionProtection     bool              `json:"enableDeletionProtection"`
	EnableHttp2                  bool              `json:"enableHttp2"`
	IdleTimeout                  int               `json:"idleTimeout"`
	IPAddressType                string            `json:"ipAddressType"`
	PreserveHostHeader           bool              `json:"preserveHostHeader"`
	EnableCrossZoneLoadBalancing string            `json:"enableCrossZoneLoadBalancing"`
	ClientKeepAlive              int               `json:"clientKeepAlive"`
	XffHeaderProcessingMode      string            `json:"xffHeaderProcessingMode"`
	Tags                         map[string]string `json:"tags"`
}

type AccessLogs struct {
	Enabled bool   `json:"enabled"`
	Bucket  string `json:"bucket,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
}

type CreateLoadBalancerInput struct {
	Name                         string            `json:"name"`
	NamePrefix                   string            `json:"namePrefix,omitempty"`
	Internal                     bool              `json:"internal"`
	Subnets                      []string          `json:"subnets"`
	SecurityGroups               []string          `json:"securityGroups,omitempty"`
	AccessLogs                   *AccessLogs       `json:"accessLogs,omitempty"`
	EnableDeletionProtection     bool              `json:"enableDeletionProtection"`
	EnableHttp2                  bool              `json:"enableHttp2"`
	IdleTimeout                  int               `json:"idleTimeout,omitempty"`
	IPAddressType                string            `json:"ipAddressType,omitempty"`
	PreserveHostHeader           bool              `json:"preserveHostHeader"`
	EnableCrossZoneLoadBalancing string            `json:"enableCrossZoneLoadBalancing,omitempty"`
	ClientKeepAlive              int               `json:"clientKeepAlive,omitempty"`
	XffHeaderProcessingMode      string            `json:"xffHeaderProcessingMode,omitempty"`
	Tags                         map[string]string `json:"tags,omitempty"`
}

func (c *Client) CreateLoadBalancer(input *CreateLoadBalancerInput) (*LoadBalancer, error) {
	resp, err := c.sendRequest("POST", fmt.Sprintf("/aws_account/%s/load-balancers", c.accountID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &lb, nil
}

func (c *Client) GetLoadBalancer(loadBalancerID string) (*LoadBalancer, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &lb, nil
}

type UpdateLoadBalancerInput struct {
    Name                         string
    SecurityGroups               []string
    AccessLogs                   *AccessLogs
    EnableDeletionProtection     *bool
    EnableHttp2                  *bool
    IdleTimeout                  *int
    PreserveHostHeader           *bool
    EnableCrossZoneLoadBalancing *string
    ClientKeepAlive              *int
    XffHeaderProcessingMode      *string
    Tags                         map[string]string
}

func (c *Client) UpdateLoadBalancer(loadBalancerID string, input *UpdateLoadBalancerInput) (*LoadBalancer, error) {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &lb, nil
}

func (c *Client) DeleteLoadBalancer(loadBalancerID string) error {
	_, err := c.sendRequest("DELETE", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), nil)
	return err
}