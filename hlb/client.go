// File: hlb/client.go

package hlb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultMaxRetries = 5
	defaultBaseURL    = "https://m2ga5q4lxe.execute-api.%s.amazonaws.com/prod/v1"
)

type Client struct {
	httpClient  *retryablehttp.Client
	baseURL     string
	apiKey      string
	awsConfig   aws.Config
	accountID   string
	credentials *Credentials
}

func NewClient(ctx context.Context, apiKey string, awsConfig aws.Config) (*Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = defaultMaxRetries
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.CheckRetry = customRetryPolicy

	credentials, err := loadOrCreateCredentials(ctx, apiKey, awsConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient:  retryClient,
		baseURL:     fmt.Sprintf(defaultBaseURL, awsConfig.Region),
		apiKey:      apiKey,
		awsConfig:   awsConfig,
		accountID:   credentials.AccountID,
		credentials: credentials,
	}, nil
}

func (c *Client) GetRegion() string {
	return c.awsConfig.Region
}

func (c *Client) GetAccountID() string {
	return c.accountID
}

func (c *Client) sendRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	
	XSTSGCIHeaders, err := getSCDIHeader(ctx, c.awsConfig, c.credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to generate API credentials: %w", err)
	}

	var buf io.Reader
	if body != nil {
		payloadBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(payloadBytes)
	}

	req, err := retryablehttp.NewRequest(method, url, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("X-Sts-Gci-Headers", XSTSGCIHeaders)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func customRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	// Retry only if the response status code is 429
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	return false, nil
}