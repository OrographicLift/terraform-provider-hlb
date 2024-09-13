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
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultMaxRetries = 5
	defaultBaseURL    = "https://lb.%s.dev.orographic.com/v1"
)

type Client struct {
	httpClient *retryablehttp.Client
	baseURL    string
	apiKey     string
	region     string
	accountID  string
}

func NewClient(apiKey string, awsConfig aws.Config) (*Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = defaultMaxRetries
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.CheckRetry = retryablehttp.DefaultRetryPolicy

	// Get the AWS account ID
	stsClient := sts.NewFromConfig(awsConfig)
	result, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("error getting AWS account ID: %v", err)
	}
	accountID := *result.Account

	return &Client{
		httpClient: retryClient,
		baseURL:    fmt.Sprintf(defaultBaseURL, awsConfig.Region),
		apiKey:     apiKey,
		region:     awsConfig.Region,
		accountID: accountID,
	}, nil
}

func (c *Client) GetRegion() string {
	return c.region
}

func (c *Client) GetAccountID() string {
	return c.accountID
}

func (c *Client) sendRequest(method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}