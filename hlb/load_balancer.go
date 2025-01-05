// File: hlb/load_balancer.go

package hlb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type LoadBalancer struct {
	AccessLogs                   *AccessLogs       `json:"accessLogs,omitempty"`
	AccountID                    string            `json:"accountId"`
	ClientKeepAlive              int               `json:"clientKeepAlive"`
	CreatedAt                    time.Time         `json:"createdAt"`
	DNSName                      string            `json:"dnsName"`
	EnableCrossZoneLoadBalancing string            `json:"enableCrossZoneLoadBalancing"`
	EnableDeletionProtection     bool              `json:"enableDeletionProtection"`
	EnableHttp2                  bool              `json:"enableHttp2"`
	ExpiresAt                    int               `json:"expiresAt"`
	ID                           string            `json:"id"`
	IdleTimeout                  int               `json:"idleTimeout"`
	Internal                     bool              `json:"internal"`
	IPAddressType                string            `json:"ipAddressType"`
	LaunchConfig                 *LaunchConfig     `json:"launchConfig"`
	Name                         string            `json:"name"`
	PreserveHostHeader           bool              `json:"preserveHostHeader"`
	SecurityGroups               []string          `json:"securityGroups"`
	State                        string            `json:"state"`
	Subnets                      []string          `json:"subnets"`
	Tags                         map[string]string `json:"tags"`
	UpdatedAt                    time.Time         `json:"updatedAt"`
	URI                          string            `json:"uri"`
	XffHeaderProcessingMode      string            `json:"xffHeaderProcessingMode"`
	ZoneID                       string            `json:"zoneId"`
	ZoneName                     string            `json:"zoneName"`
}

type LaunchConfig struct {
	InstanceType     string `json:"instanceType"`
	MinInstanceCount int    `json:"minInstanceCount"`
	MaxInstanceCount int    `json:"maxInstanceCount"`
	TargetCPUUsage   int    `json:"targetCpuUsage"`
}

type AccessLogs struct {
	Bucket  string `json:"bucket"`
	Enabled bool   `json:"enabled"`
	Prefix  string `json:"prefix"`
}

type LoadBalancerCreate struct {
	AccessLogs                   *AccessLogs       `json:"accessLogs"`
	ClientKeepAlive              int               `json:"clientKeepAlive"`
	EnableCrossZoneLoadBalancing string            `json:"enableCrossZoneLoadBalancing"`
	EnableDeletionProtection     bool              `json:"enableDeletionProtection"`
	EnableHttp2                  bool              `json:"enableHttp2"`
	IdleTimeout                  int               `json:"idleTimeout"`
	Internal                     bool              `json:"internal"`
	IPAddressType                string            `json:"ipAddressType"`
	LaunchConfig                 *LaunchConfig     `json:"launchConfig"`
	Name                         string            `json:"name"`
	NamePrefix                   string            `json:"namePrefix"`
	PreserveHostHeader           bool              `json:"preserveHostHeader"`
	SecurityGroups               []string          `json:"securityGroups"`
	Subnets                      []string          `json:"subnets"`
	Tags                         map[string]string `json:"tags"`
	XffHeaderProcessingMode      string            `json:"xffHeaderProcessingMode"`
	ZoneID                       string            `json:"zoneId"`
	ZoneName                     string            `json:"zoneName"`
}

type LoadBalancerUpdate struct {
	AccessLogs                   *AccessLogs        `json:"accessLogs"`
	ClientKeepAlive              *int               `json:"clientKeepAlive"`
	EnableCrossZoneLoadBalancing *string            `json:"enableCrossZoneLoadBalancing"`
	EnableDeletionProtection     *bool              `json:"enableDeletionProtection"`
	EnableHttp2                  *bool              `json:"enableHttp2"`
	IdleTimeout                  *int               `json:"idleTimeout"`
	LaunchConfig                 *LaunchConfig      `json:"launchConfig"`
	Name                         *string            `json:"name"`
	PreserveHostHeader           *bool              `json:"preserveHostHeader"`
	SecurityGroups               []string           `json:"securityGroups"`
	Tags                         *map[string]string `json:"tags"`
	XffHeaderProcessingMode      *string            `json:"xffHeaderProcessingMode"`
}

const (
	LBStatePendingCreation = "pending_creation"
	LBStateCreating        = "creating"
	LBStatePendingUpdate   = "pending_update"
	LBStateUpdating        = "updating"
	LBStatePendingDeletion = "pending_delete"
	LBStateDeleting        = "deleting"
	LBStateDeleted         = "deleted"
	LBStateActive          = "active"
	LBStateFailed          = "failed"

	// Default timeouts
	DefaultCreateTimeout = 30 * time.Minute
	DefaultUpdateTimeout = 30 * time.Minute
	DefaultDeleteTimeout = 30 * time.Minute
)

func isLoadBalancerInPendingState(state string) bool {
	pendingStates := map[string]bool{
		LBStatePendingCreation: true,
		LBStateCreating:        true,
		LBStatePendingUpdate:   true,
		LBStateUpdating:        true,
		LBStatePendingDeletion: true,
		LBStateDeleting:        true,
	}
	return pendingStates[state]
}

func (c *Client) waitForLoadBalancerState(ctx context.Context, id string, target []string, timeout time.Duration) (*LoadBalancer, error) {
	targetStates := make(map[string]bool)
	for _, s := range target {
		targetStates[s] = true
	}

	var lb *LoadBalancer
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		var err error
		lb, err = c.GetLoadBalancer(ctx, id)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("error getting load balancer (%s): %w", id, err))
		}

		if lb.State == LBStateFailed {
			return retry.NonRetryableError(fmt.Errorf("load balancer (%s) entered failed state", id))
		}

		if targetStates[lb.State] {
			return nil
		}

		if isLoadBalancerInPendingState(lb.State) {
			return retry.RetryableError(fmt.Errorf("expected load balancer (%s) to be in state %v but was in state %s", id, target, lb.State))
		}

		return retry.NonRetryableError(fmt.Errorf("load balancer (%s) entered unexpected state %s", id, lb.State))
	})

	if err != nil {
		return nil, err
	}

	return lb, nil
}

func (c *Client) ListLoadBalancers(ctx context.Context, limit int, nextToken string) ([]LoadBalancer, string, error) {
	url := fmt.Sprintf("/aws_account/%s/load-balancers?limit=%d", c.accountID, limit)
	if nextToken != "" {
		url += fmt.Sprintf("&nextToken=%s", nextToken)
	}

	resp, err := c.sendRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var response struct {
		Items     []LoadBalancer `json:"items"`
		NextToken string         `json:"nextToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Items, response.NextToken, nil
}

func (c *Client) CreateLoadBalancer(ctx context.Context, input *LoadBalancerCreate) (*LoadBalancer, error) {
	resp, err := c.sendRequest(ctx, "POST", fmt.Sprintf("/aws_account/%s/load-balancers", c.accountID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Wait for the load balancer to be active
	return c.waitForLoadBalancerState(ctx, lb.ID, []string{LBStateActive}, DefaultCreateTimeout)
}

func (c *Client) GetLoadBalancer(ctx context.Context, loadBalancerID string) (*LoadBalancer, error) {
	resp, err := c.sendRequest(ctx, "GET", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), nil)
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

func (c *Client) UpdateLoadBalancer(ctx context.Context, loadBalancerID string, input *LoadBalancerUpdate) (*LoadBalancer, error) {
	resp, err := c.sendRequest(ctx, "PUT", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var lb LoadBalancer
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Wait for the load balancer to be active after update
	return c.waitForLoadBalancerState(ctx, lb.ID, []string{LBStateActive}, DefaultUpdateTimeout)
}

func (c *Client) DeleteLoadBalancer(ctx context.Context, loadBalancerID string) error {
	resp, err := c.sendRequest(ctx, "DELETE", fmt.Sprintf("/aws_account/%s/load-balancers/%s", c.accountID, loadBalancerID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Wait for the load balancer to be deleted
	_, err = c.waitForLoadBalancerState(ctx, loadBalancerID, []string{LBStateDeleted}, DefaultDeleteTimeout)
	return err
}
