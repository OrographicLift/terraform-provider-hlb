// File: client.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	APIKey  string
	BaseURL string
	HTTP    *retryablehttp.Client
}

type HLB struct {
	ID             string                 `json:"id"`
	TargetGroupARN string                 `json:"targetGroupArn"`
	FQDN           string                 `json:"fqdn"`
	Route53ZoneID  string                 `json:"route53ZoneId"`
	Status         string                 `json:"status"`
	Config         map[string]interface{} `json:"config"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

func NewClient(apiKey, baseURL string) *Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy

	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTP:    retryClient,
	}
}

func (c *Client) sendRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		payloadBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling request body: %w", err)
		}
		buf = bytes.NewBuffer(payloadBytes)
	}

	req, err := retryablehttp.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, path), buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request error: %s (status code: %d)", string(body), resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) CreateHLB(hlb *HLB) (*HLB, error) {
	resp, err := c.sendRequest("POST", "/hlb", hlb)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var createdHLB HLB
	if err := json.NewDecoder(resp.Body).Decode(&createdHLB); err != nil {
		return nil, err
	}

	return &createdHLB, nil
}

func (c *Client) GetHLB(id string) (*HLB, error) {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/hlb/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hlb HLB
	if err := json.NewDecoder(resp.Body).Decode(&hlb); err != nil {
		return nil, err
	}

	return &hlb, nil
}

func (c *Client) UpdateHLB(id string, hlb *HLB) (*HLB, error) {
	resp, err := c.sendRequest("PUT", fmt.Sprintf("/hlb/%s", id), hlb)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var updatedHLB HLB
	if err := json.NewDecoder(resp.Body).Decode(&updatedHLB); err != nil {
		return nil, err
	}

	return &updatedHLB, nil
}

func (c *Client) DeleteHLB(id string) error {
	_, err := c.sendRequest("DELETE", fmt.Sprintf("/hlb/%s", id), nil)
	return err
}
