package cmd

import (
	"fmt"
	"os"

	"aws-switcher/internal/aws"
	"aws-switcher/internal/config"
	"aws-switcher/internal/ui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aws-switcher",
	Short: "AWS SSO credential switcher",
	Long:  "A tool to switch between AWS SSO accounts and update credentials",
	Run:   runRoot,
}

var (
	startURL  string
	ssoRegion string
	update    bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&startURL, "start-url", "", "AWS SSO start URL")
	rootCmd.PersistentFlags().StringVar(&ssoRegion, "sso-region", "us-east-1", "AWS SSO region")
	rootCmd.PersistentFlags().BoolVar(&update, "update", false, "Update account list from SSO")
}

func Execute() error {
	return rootCmd.Execute()
}

func runRoot(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// If update flag is set or no accounts saved, fetch from SSO
	if update || len(cfg.Accounts) == 0 {
		if startURL == "" {
			startURL = cfg.StartURL
		}
		if startURL == "" {
			fmt.Fprintf(os.Stderr, "SSO start URL is required. Use --start-url flag or save in config\n")
			os.Exit(1)
		}

		fmt.Println("Fetching available accounts from SSO...")
		accounts, err := aws.FetchAccounts(startURL, ssoRegion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching accounts: %v\n", err)
			os.Exit(1)
		}

		cfg.StartURL = startURL
		cfg.SSORegion = ssoRegion
		cfg.Accounts = accounts

		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated %d accounts\n", len(accounts))
		return
	}

	// Interactive account selection
	if len(cfg.Accounts) == 0 {
		fmt.Println("No accounts found. Run with --update flag to fetch accounts from SSO.")
		return
	}

	selectedAccount, err := ui.SelectAccount(cfg.Accounts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting account: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected: %s (%s) - %s\n", selectedAccount.Name, selectedAccount.ID, selectedAccount.RoleName)

	// Login and get credentials
	fmt.Println("Logging in to AWS SSO...")
	creds, err := aws.GetCredentials(cfg.StartURL, cfg.SSORegion, selectedAccount.ID, selectedAccount.RoleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting credentials: %v\n", err)
		os.Exit(1)
	}

	// Write credentials to AWS config
	if err := aws.WriteCredentials(creds); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing credentials: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated AWS credentials!")
}
