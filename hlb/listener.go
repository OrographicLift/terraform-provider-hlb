package hlb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Listener struct {
	ID                        string    `json:"id"`
	URI                       string    `json:"uri"`
	LoadBalancerID            string    `json:"loadBalancerId"`
	Port                      int       `json:"port"`
	Protocol                  string    `json:"protocol"`
	TargetGroupARN            string    `json:"targetGroupArn"`
	CertificateSecretsName    string    `json:"certificateSecretsName,omitempty"`
	ALPNPolicy                string    `json:"alpnPolicy,omitempty"`
	EnableDeletionProtection  bool      `json:"enableDeletionProtection"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

type CreateListenerInput struct {
	Port                      int    `json:"port"`
	Protocol                  string `json:"protocol"`
	TargetGroupARN            string `json:"targetGroupArn"`
	CertificateSecretsName    string `json:"certificateSecretsName,omitempty"`
	ALPNPolicy                string `json:"alpnPolicy,omitempty"`
	EnableDeletionProtection  bool   `json:"enableDeletionProtection"`
}

type UpdateListenerInput struct {
	Port                      *int    `json:"port,omitempty"`
	Protocol                  *string `json:"protocol,omitempty"`
	TargetGroupARN            *string `json:"targetGroupArn,omitempty"`
	CertificateSecretsName    *string `json:"certificateSecretsName,omitempty"`
	ALPNPolicy                *string `json:"alpnPolicy,omitempty"`
	EnableDeletionProtection  *bool   `json:"enableDeletionProtection"`
}

func (c *Client) CreateListener(ctx context.Context, loadBalancerID string, input *CreateListenerInput) (*Listener, error) {
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

func (c *Client) UpdateListener(ctx context.Context, loadBalancerID, listenerID string, input *UpdateListenerInput) (*Listener, error) {
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
	url := fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners?limit=%d", c.accountID, loadBalancerID, limit)
	if nextToken != "" {
		url += fmt.Sprintf("&nextToken=%s", nextToken)
	}

	resp, err := c.sendRequest(ctx, "GET", url, nil)
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