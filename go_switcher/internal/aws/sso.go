package aws

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aws-switcher/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/pkg/browser"
	"gopkg.in/ini.v1"
)

// Credentials represents AWS temporary credentials
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// FetchAccounts fetches available accounts from AWS SSO
func FetchAccounts(startURL, ssoRegion string) ([]config.Account, error) {
	ctx := context.Background()

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(ssoRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Get access token
	accessToken, err := getAccessToken(ctx, cfg, startURL, ssoRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create SSO client
	ssoClient := sso.NewFromConfig(cfg)

	// List accounts
	accountsResp, err := ssoClient.ListAccounts(ctx, &sso.ListAccountsInput{
		AccessToken: aws.String(accessToken),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	var accounts []config.Account
	for _, account := range accountsResp.AccountList {
		// List roles for each account
		rolesResp, err := ssoClient.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
			AccessToken: aws.String(accessToken),
			AccountId:   account.AccountId,
		})
		if err != nil {
			continue // skip accounts we can't access
		}

		for _, role := range rolesResp.RoleList {
			accounts = append(accounts, config.Account{
				ID:       aws.ToString(account.AccountId),
				Name:     aws.ToString(account.AccountName),
				RoleName: aws.ToString(role.RoleName),
			})
		}
	}

	return accounts, nil
}

// GetCredentials gets temporary credentials for the specified account and role
func GetCredentials(startURL, ssoRegion, accountID, roleName string) (*Credentials, error) {
	ctx := context.Background()

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(ssoRegion))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Get access token
	accessToken, err := getAccessToken(ctx, cfg, startURL, ssoRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create SSO client
	ssoClient := sso.NewFromConfig(cfg)

	// Get role credentials
	credsResp, err := ssoClient.GetRoleCredentials(ctx, &sso.GetRoleCredentialsInput{
		AccessToken: aws.String(accessToken),
		AccountId:   aws.String(accountID),
		RoleName:    aws.String(roleName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get role credentials: %w", err)
	}

	return &Credentials{
		AccessKeyID:     aws.ToString(credsResp.RoleCredentials.AccessKeyId),
		SecretAccessKey: aws.ToString(credsResp.RoleCredentials.SecretAccessKey),
		SessionToken:    aws.ToString(credsResp.RoleCredentials.SessionToken),
		Expiration:      time.Unix(credsResp.RoleCredentials.Expiration/1000, 0),
	}, nil
}

// getAccessToken performs the SSO OIDC flow to get an access token
func getAccessToken(ctx context.Context, cfg aws.Config, startURL, ssoRegion string) (string, error) {
	// Create SSO OIDC client
	oidcClient := ssooidc.NewFromConfig(cfg)

	// Register client
	registerResp, err := oidcClient.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String("aws-switcher"),
		ClientType: aws.String("public"),
		Scopes:     []string{"sso:account:access"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to register client: %w", err)
	}

	// Start device authorization
	authResp, err := oidcClient.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     registerResp.ClientId,
		ClientSecret: registerResp.ClientSecret,
		StartUrl:     aws.String(startURL),
	})
	if err != nil {
		return "", fmt.Errorf("failed to start device authorization: %w", err)
	}

	// Get the complete verification URI
	verificationURL := aws.ToString(authResp.VerificationUriComplete)
	userCode := aws.ToString(authResp.UserCode)

	// Display information to user
	fmt.Printf("Opening browser for AWS SSO authorization...\n")
	fmt.Printf("If the browser doesn't open automatically, please visit:\n")
	fmt.Printf("URL: %s\n", verificationURL)
	fmt.Printf("Code: %s\n", userCode)

	// Attempt to open browser automatically
	if err := browser.OpenURL(verificationURL); err != nil {
		fmt.Printf("Failed to open browser automatically: %v\n", err)
		fmt.Printf("Please open the URL manually in your browser.\n")
	} else {
		fmt.Printf("Browser opened successfully!\n")
	}

	fmt.Printf("Waiting for authorization...\n")

	// Poll for token
	tokenInput := &ssooidc.CreateTokenInput{
		ClientId:     registerResp.ClientId,
		ClientSecret: registerResp.ClientSecret,
		DeviceCode:   authResp.DeviceCode,
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
	}

	expiresAt := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	interval := time.Duration(authResp.Interval) * time.Second

	for time.Now().Before(expiresAt) {
		tokenResp, err := oidcClient.CreateToken(ctx, tokenInput)
		if err != nil {
			var pendingAuth *types.AuthorizationPendingException
			var slowDown *types.SlowDownException

			if errors.As(err, &pendingAuth) {
				time.Sleep(interval)
				continue
			}
			if errors.As(err, &slowDown) {
				interval += 5 * time.Second
				time.Sleep(interval)
				continue
			}
			return "", fmt.Errorf("failed to create token: %w", err)
		}

		return aws.ToString(tokenResp.AccessToken), nil
	}

	return "", fmt.Errorf("authorization timed out")
}

// WriteCredentials writes AWS credentials to the default profile
func WriteCredentials(creds *Credentials) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	awsDir := filepath.Join(homeDir, ".aws")
	if err := os.MkdirAll(awsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .aws directory: %w", err)
	}

	credentialsPath := filepath.Join(awsDir, "credentials")

	// Load existing credentials file or create new one
	var cfg *ini.File
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		cfg = ini.Empty()
	} else {
		cfg, err = ini.Load(credentialsPath)
		if err != nil {
			return fmt.Errorf("failed to load credentials file: %w", err)
		}
	}

	// Update default section
	section, err := cfg.NewSection("default")
	if err != nil {
		// Section already exists, get it
		section = cfg.Section("default")
	}

	section.Key("aws_access_key_id").SetValue(creds.AccessKeyID)
	section.Key("aws_secret_access_key").SetValue(creds.SecretAccessKey)
	section.Key("aws_session_token").SetValue(creds.SessionToken)

	// Save the file
	if err := cfg.SaveTo(credentialsPath); err != nil {
		return fmt.Errorf("failed to save credentials file: %w", err)
	}

	// Ensure config file exists with default profile
	configPath := filepath.Join(awsDir, "config")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configCfg := ini.Empty()
		configCfg.NewSection("profile default")
		if err := configCfg.SaveTo(configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	return nil
}
