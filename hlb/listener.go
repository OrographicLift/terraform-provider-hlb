package hlb

import (
	"encoding/json"
	"fmt"
	"time"
)

type Listener struct {
	ID                    string    `json:"id"`
	URI                   string    `json:"uri"`
	LoadBalancerID        string    `json:"loadBalancerId"`
	Port                  int       `json:"port"`
	Protocol              string    `json:"protocol"`
	TargetGroupARN        string    `json:"targetGroupArn"`
	CertificateSecretsARN string    `json:"certificateSecretsArn,omitempty"`
	ALPNPolicy            string    `json:"alpnPolicy,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type CreateListenerInput struct {
	Port                  int    `json:"port"`
	Protocol              string `json:"protocol"`
	TargetGroupARN        string `json:"targetGroupArn"`
	CertificateSecretsARN string `json:"certificateSecretsArn,omitempty"`
	ALPNPolicy            string `json:"alpnPolicy,omitempty"`
}

type UpdateListenerInput struct {
	Port                  *int    `json:"port,omitempty"`
	Protocol              *string `json:"protocol,omitempty"`
	TargetGroupARN        *string `json:"targetGroupArn,omitempty"`
	CertificateSecretsARN *string `json:"certificateSecretsArn,omitempty"`
	ALPNPolicy            *string `json:"alpnPolicy,omitempty"`
}

func (c *Client) CreateListener(loadBalancerID string, input *CreateListenerInput) (*Listener, error) {
	resp, err := c.sendRequest("POST", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners", c.accountID, loadBalancerID), input)
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

func (c *Client) GetListener(loadBalancerID, listenerID string) (*Listener, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), nil)
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

func (c *Client) UpdateListener(loadBalancerID, listenerID string, input *UpdateListenerInput) (*Listener, error) {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), input)
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

func (c *Client) DeleteListener(loadBalancerID, listenerID string) error {
	_, err := c.sendRequest("DELETE", fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners/%s", c.accountID, loadBalancerID, listenerID), nil)
	return err
}

func (c *Client) ListListeners(loadBalancerID string, limit int, nextToken string) ([]Listener, string, error) {
	url := fmt.Sprintf("/aws_account/%s/load-balancers/%s/listeners?limit=%d", c.accountID, loadBalancerID, limit)
	if nextToken != "" {
		url += fmt.Sprintf("&nextToken=%s", nextToken)
	}

	resp, err := c.sendRequest("GET", url, nil)
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