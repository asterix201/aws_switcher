package config

import (
	"encoding/json"
	"testing"
)

func TestConfigMarshalUnmarshal(t *testing.T) {
	// Test config with sample data
	testConfig := &Config{
		StartURL:  "https://test.awsapps.com/start/",
		SSORegion: "us-west-2",
		Accounts: []Account{
			{
				ID:       "123456789012",
				Name:     "Test Account",
				RoleName: "TestRole",
			},
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal from JSON
	var loadedConfig Config
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify loaded config matches original config
	if loadedConfig.StartURL != testConfig.StartURL {
		t.Errorf("StartURL mismatch: got %s, want %s", loadedConfig.StartURL, testConfig.StartURL)
	}

	if loadedConfig.SSORegion != testConfig.SSORegion {
		t.Errorf("SSORegion mismatch: got %s, want %s", loadedConfig.SSORegion, testConfig.SSORegion)
	}

	if len(loadedConfig.Accounts) != len(testConfig.Accounts) {
		t.Errorf("Accounts length mismatch: got %d, want %d", len(loadedConfig.Accounts), len(testConfig.Accounts))
	}

	if len(loadedConfig.Accounts) > 0 && len(testConfig.Accounts) > 0 {
		if loadedConfig.Accounts[0].ID != testConfig.Accounts[0].ID {
			t.Errorf("Account ID mismatch: got %s, want %s", loadedConfig.Accounts[0].ID, testConfig.Accounts[0].ID)
		}
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Should return default values when no config file exists
	if cfg.SSORegion != "us-east-1" {
		t.Errorf("Expected default SSORegion 'us-east-1', got '%s'", cfg.SSORegion)
	}

	if cfg.Accounts == nil {
		t.Error("Expected Accounts to be initialized, got nil")
	}
}

func TestAccount(t *testing.T) {
	account := Account{
		ID:       "123456789012",
		Name:     "Test Account",
		RoleName: "AdminRole",
	}

	if account.ID == "" {
		t.Error("Account ID should not be empty")
	}

	if account.Name == "" {
		t.Error("Account Name should not be empty")
	}

	if account.RoleName == "" {
		t.Error("Account RoleName should not be empty")
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}

	// Test that we can set defaults
	if cfg.SSORegion == "" {
		cfg.SSORegion = "us-east-1"
	}

	if cfg.Accounts == nil {
		cfg.Accounts = make([]Account, 0)
	}

	if cfg.SSORegion != "us-east-1" {
		t.Errorf("Expected SSORegion to be set to 'us-east-1', got '%s'", cfg.SSORegion)
	}

	if cfg.Accounts == nil {
		t.Error("Expected Accounts to be initialized")
	}
}

func TestMultipleAccounts(t *testing.T) {
	accounts := []Account{
		{ID: "111111111111", Name: "Account 1", RoleName: "Role1"},
		{ID: "222222222222", Name: "Account 2", RoleName: "Role2"},
		{ID: "333333333333", Name: "Account 3", RoleName: "Role3"},
	}

	cfg := &Config{
		StartURL:  "https://test.awsapps.com/start/",
		SSORegion: "us-west-2",
		Accounts:  accounts,
	}

	// Test JSON marshaling with multiple accounts
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config with multiple accounts: %v", err)
	}

	// Test JSON unmarshaling
	var loadedConfig Config
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("Failed to unmarshal config with multiple accounts: %v", err)
	}

	if len(loadedConfig.Accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(loadedConfig.Accounts))
	}

	for i, account := range loadedConfig.Accounts {
		if account.ID != accounts[i].ID {
			t.Errorf("Account %d ID mismatch: got %s, want %s", i, account.ID, accounts[i].ID)
		}
	}
}
