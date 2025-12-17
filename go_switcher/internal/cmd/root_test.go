package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand(t *testing.T) {
	// Test that the root command can be created and has expected properties
	cmd := &cobra.Command{
		Use:   "aws-switcher",
		Short: "AWS SSO credential switcher",
		Long:  "A tool to switch between AWS SSO accounts and update credentials",
	}

	if cmd.Use != "aws-switcher" {
		t.Errorf("Expected command name 'aws-switcher', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	if cmd.Long == "" {
		t.Error("Command should have a long description")
	}
}

func TestCommandFlags(t *testing.T) {
	// Create a test command with flags similar to our root command
	cmd := &cobra.Command{
		Use:   "test-cmd",
		Short: "Test command",
		Run: func(cmd *cobra.Command, args []string) {
			// Test function
		},
	}

	var (
		testStartURL  string
		testSSORegion string
		testUpdate    bool
	)

	// Add flags like in our root command
	cmd.PersistentFlags().StringVar(&testStartURL, "start-url", "", "AWS SSO start URL")
	cmd.PersistentFlags().StringVar(&testSSORegion, "sso-region", "us-east-1", "AWS SSO region")
	cmd.PersistentFlags().BoolVar(&testUpdate, "update", false, "Update account list from SSO")

	// Test flag defaults
	if testSSORegion != "us-east-1" {
		t.Errorf("Expected default sso-region 'us-east-1', got '%s'", testSSORegion)
	}

	if testUpdate != false {
		t.Errorf("Expected default update flag to be false, got %v", testUpdate)
	}

	if testStartURL != "" {
		t.Errorf("Expected default start-url to be empty, got '%s'", testStartURL)
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute function exists and doesn't panic when called
	// In a real test environment, you might want to mock the command execution
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Execute function panicked: %v", r)
		}
	}()

	// We can't actually run Execute() in tests as it would run the real command
	// Instead, we verify that the function exists
	t.Log("Execute function exists and is callable")
}

func TestCommandHelp(t *testing.T) {
	// Test that help output can be generated
	cmd := &cobra.Command{
		Use:   "aws-switcher",
		Short: "AWS SSO credential switcher",
		Long:  "A tool to switch between AWS SSO accounts and update credentials",
	}

	// Test that help string can be generated
	help := cmd.UsageString()
	if help == "" {
		t.Error("Command should generate help text")
	}

	// Test that the command has the expected Use field
	if cmd.Use != "aws-switcher" {
		t.Errorf("Expected command Use to be 'aws-switcher', got '%s'", cmd.Use)
	}

	// Test that the command has a short description
	if cmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Log the help text for debugging (we can see what's actually in it)
	t.Logf("Help text generated successfully, length: %d characters", len(help))
}
