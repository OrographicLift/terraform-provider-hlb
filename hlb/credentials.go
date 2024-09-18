package hlb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ini "gopkg.in/ini.v1"
)

const (
	credentialsDir  = ".hlb"
	credentialsFile = "credentials"
	defaultSection  = "default"
	expiryDuration  = 15 * time.Minute
)

type Credentials struct {
	APIKey           string
	XSTSGCIHeaders   string
	Expiry           time.Time
	AccountID        string
}

func getSCDIHeader(ctx context.Context, cfg aws.Config, credentials *Credentials) (string, error) {
	if time.Now().After(credentials.Expiry) {
		headers, err := generateSTSHeaders(ctx, cfg)
		if err != nil {
			return "", fmt.Errorf("failed to generate STS headers: %w", err)
		}

		credentials.XSTSGCIHeaders = headers
		credentials.Expiry = time.Now().Add(expiryDuration)

		if err := saveCredentials(credentials); err != nil {
			return "", fmt.Errorf("failed to save credentials: %w", err)
		}
	}
	return credentials.XSTSGCIHeaders, nil
}

func loadOrCreateCredentials(ctx context.Context, apiKey string, cfg aws.Config) (*Credentials, error) {
	var credentials *Credentials
	var accountID string
	credentials, err := loadCredentials(apiKey)
	if err != nil {
		return nil, err
	}
	if credentials == nil {
		// Get the AWS account ID
		stsClient := sts.NewFromConfig(cfg)
		result, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
		if err != nil {
			return nil, fmt.Errorf("error getting AWS account ID: %v", err)
		}
		accountID = *result.Account

		headers, err := generateSTSHeaders(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate STS headers: %w", err)
		}
		credentials = &Credentials{
			APIKey:         apiKey,
			XSTSGCIHeaders: headers,
			Expiry:         time.Now().Add(expiryDuration),
			AccountID:      accountID,
		}
		if err := saveCredentials(credentials); err != nil {
			return nil, fmt.Errorf("failed to save credentials: %w", err)
		}
	}
	return credentials, nil
}

func loadCredentials(apiKey string) (*Credentials, error) {
	credPath := getCredentialsPath()
	cfg, err := ini.Load(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load credentials file: %w", err)
	}

	section := cfg.Section(apiKey)
	if section == nil || section.Key("account_id").String() == "" || section.Key("x_sts_gci_headers").String() == "" {
		return nil, nil
	}

	expiry, _ := time.Parse(time.RFC3339, section.Key("expiry").String())
	return &Credentials{
		APIKey:         apiKey,
		XSTSGCIHeaders: section.Key("x_sts_gci_headers").String(),
		Expiry:         expiry,
		AccountID:      section.Key("account_id").String(),
	}, nil
}

func saveCredentials(creds *Credentials) error {
	credPath := getCredentialsPath()
	cfg, err := ini.Load(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = ini.Empty()
		} else {
			return fmt.Errorf("failed to load credentials file: %w", err)
		}
	}

	section, err := cfg.NewSection(creds.APIKey)
	if err != nil {
		return fmt.Errorf("failed to create section in credentials file: %w", err)
	}

	section.NewKey("x_sts_gci_headers", creds.XSTSGCIHeaders)
	section.NewKey("expiry", creds.Expiry.Format(time.RFC3339))
	section.NewKey("account_id", creds.AccountID)

	if err := ensureCredentialsDir(); err != nil {
		return err
	}
	if err := cfg.SaveTo(credPath); err != nil {
		return fmt.Errorf("failed to save credentials file: %w", err)
	}

	return nil
}

func getCredentialsPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, credentialsDir, credentialsFile)
}

func generateSTSHeaders(ctx context.Context, cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)

	presigner := sts.NewPresignClient(stsClient)
	presignedURL, err := presigner.PresignGetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}, func(opts *sts.PresignOptions) {})
	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}

	// Extract query string from presigned URL
	parsedURL, err := url.Parse(presignedURL.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse presigned URL: %w", err)
	}

	headers := parsedURL.Query().Encode()

	return headers, nil
}

func ensureCredentialsDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	credDir := filepath.Join(homeDir, credentialsDir)
	if err := os.MkdirAll(credDir, 0700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	credPath := filepath.Join(credDir, credentialsFile)
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		file, err := os.OpenFile(credPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to create credentials file: %w", err)
		}
		file.Close()
	}

	return nil
}