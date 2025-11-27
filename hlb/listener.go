package hlb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Listener struct {
	ALPNPolicy               string    `json:"alpnPolicy,omitempty"`
	CertificateSecretsName   string    `json:"certificateSecretsName,omitempty"`
	CreatedAt                time.Time `json:"createdAt"`
	EnableDeletionProtection bool      `json:"enableDeletionProtection"`
	ID                       string    `json:"id"`
	LoadBalancerID           string    `json:"loadBalancerId"`
	OverprovisioningFactor   float64   `json:"overprovisioningFactor"`
	Port                     int       `json:"port"`
	Protocol                 string    `json:"protocol"`
	TargetGroupARN           string    `json:"targetGroupArn"`
	UpdatedAt                time.Time `json:"updatedAt"`
	URI                      string    `json:"uri"`
}

type ListenerCreate struct {
	ALPNPolicy               string  `json:"alpnPolicy,omitempty"`
	CertificateSecretsName   string  `json:"certificateSecretsName,omitempty"`
	EnableDeletionProtection bool    `json:"enableDeletionProtection"`
	OverprovisioningFactor   float64 `json:"overprovisioningFactor"`
	Port                     int     `json:"port"`
	Protocol                 string  `json:"protocol"`
	TargetGroupARN           string  `json:"targetGroupArn"`
}

type ListenerUpdate struct {
	ALPNPolicy               *string  `json:"alpnPolicy,omitempty"`
	CertificateSecretsName   *string  `json:"certificateSecretsName,omitempty"`
	EnableDeletionProtection *bool    `json:"enableDeletionProtection"`
	OverprovisioningFactor   *float64 `json:"overprovisioningFactor"`
	Port                     *int     `json:"port,omitempty"`
	Protocol                 *string  `json:"protocol,omitempty"`
	TargetGroupARN           *string  `json:"targetGroupArn,omitempty"`
}

func (c *Client) CreateListener(ctx context.Context, loadBalancerID string, input *ListenerCreate) (*Listener, error) {
	resp, err := c.sendRequest(ctx, "POST", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners", c.accountID, loadBalancerID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listener Listener
	if err := json.NewDecoder(resp.Body).Decode(&listener); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &listener, nil
}

func (c *Client) GetListener(ctx context.Context, loadBalancerID, listenerID string) (*Listener, error) {
	resp, err := c.sendRequest(ctx, "GET", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listener Listener
	if err := json.NewDecoder(resp.Body).Decode(&listener); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &listener, nil
}

func (c *Client) UpdateListener(ctx context.Context, loadBalancerID, listenerID string, input *ListenerUpdate) (*Listener, error) {
	resp, err := c.sendRequest(ctx, "PUT", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), input)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listener Listener
	if err := json.NewDecoder(resp.Body).Decode(&listener); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &listener, nil
}

func (c *Client) DeleteListener(ctx context.Context, loadBalancerID, listenerID string) error {
	resp, err := c.sendRequest(ctx, "DELETE", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func (c *Client) ListListeners(ctx context.Context, loadBalancerID string, limit int, nextToken string) ([]Listener, string, error) {
	urlPath := fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners?limit=%d", c.accountID, loadBalancerID, limit)
	if nextToken != "" {
		urlPath += fmt.Sprintf("&nextToken=%s", url.QueryEscape(nextToken))
	}

	resp, err := c.sendRequest(ctx, "GET", urlPath, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var response struct {
		Items     []Listener `json:"items"`
		NextToken string     `json:"nextToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Items, response.NextToken, nil
}
