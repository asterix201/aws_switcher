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
	Use:   "aws-switcher [account-id role-name]",
	Short: "AWS SSO credential switcher",
	Long: `A tool to switch between AWS SSO accounts and update credentials.

Run without arguments for interactive selection, or pass an account ID and a
role name to switch directly without the picker (useful for automation), e.g.:

  aws-switcher 058264159440 psc-read-only-secret-perm-set

The SSO browser confirmation / MFA still applies in direct mode.`,
	Args: cobra.MaximumNArgs(2),
	Run:  runRoot,
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

	// Resolve effective SSO settings: an explicit flag wins, otherwise fall back
	// to the saved config.
	if !cmd.Flags().Changed("start-url") && cfg.StartURL != "" {
		startURL = cfg.StartURL
	}
	if !cmd.Flags().Changed("sso-region") && cfg.SSORegion != "" {
		ssoRegion = cfg.SSORegion
	}

	// Explicit update request: fetch and persist the account list, then exit.
	if update {
		if err := updateAccounts(cfg, nil); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	// Direct (non-interactive) mode: account ID + role name passed as args.
	if len(args) == 2 {
		// Warm cache: validate the pair up front so an unavailable account/role
		// fails before we ever open the browser.
		if len(cfg.Accounts) > 0 {
			account, err := resolveAccount(cfg, args[0], args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			switchToAccount(cfg, nil, account)
			return
		}

		// Cold cache: authenticate once, then reuse the same session to fetch the
		// account list and to request credentials (single browser confirmation).
		session, err := login()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if err := updateAccounts(cfg, session); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		account, err := resolveAccount(cfg, args[0], args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		switchToAccount(cfg, session, account)
		return
	}
	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "Both account ID and role name are required for direct mode, e.g.:\n  aws-switcher 058264159440 psc-read-only-secret-perm-set\n")
		os.Exit(1)
	}

	// No accounts cached yet: fetch them first.
	if len(cfg.Accounts) == 0 {
		if err := updateAccounts(cfg, nil); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive account selection.
	selectedAccount, err := ui.SelectAccount(cfg.Accounts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting account: %v\n", err)
		os.Exit(1)
	}

	switchToAccount(cfg, nil, selectedAccount)
}

// login authenticates with AWS SSO and returns a session whose access token is
// reused across calls, so the browser confirmation happens only once.
func login() (*aws.Session, error) {
	if startURL == "" {
		return nil, fmt.Errorf("SSO start URL is required. Use --start-url flag or save in config")
	}

	fmt.Println("Logging in to AWS SSO...")
	session, err := aws.Login(startURL, ssoRegion)
	if err != nil {
		return nil, fmt.Errorf("error logging in: %w", err)
	}
	return session, nil
}

// updateAccounts fetches the available accounts from SSO and persists them to
// the config file. When session is nil it authenticates first.
func updateAccounts(cfg *config.Config, session *aws.Session) error {
	if session == nil {
		var err error
		if session, err = login(); err != nil {
			return err
		}
	}

	fmt.Println("Fetching available accounts from SSO...")
	accounts, err := session.FetchAccounts()
	if err != nil {
		return fmt.Errorf("error fetching accounts: %w", err)
	}

	cfg.StartURL = startURL
	cfg.SSORegion = ssoRegion
	cfg.Accounts = accounts

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	fmt.Printf("Successfully updated %d accounts\n", len(accounts))
	return nil
}

// resolveAccount looks up the requested account ID / role name in the cached
// account list. When the list is populated but the pair is missing, it returns
// an error so we fail fast before triggering the browser flow. When no accounts
// are cached yet, it falls back to the provided values and lets SSO validate
// them during the credential request.
func resolveAccount(cfg *config.Config, accountID, roleName string) (*config.Account, error) {
	for i := range cfg.Accounts {
		if cfg.Accounts[i].ID == accountID && cfg.Accounts[i].RoleName == roleName {
			return &cfg.Accounts[i], nil
		}
	}

	if len(cfg.Accounts) > 0 {
		return nil, fmt.Errorf("account %q with role %q is not available; run with --update to refresh the account list", accountID, roleName)
	}

	// Nothing cached to validate against; SSO will reject it if it is not accessible.
	return &config.Account{ID: accountID, Name: accountID, RoleName: roleName}, nil
}

// switchToAccount writes temporary credentials for the given account/role. When
// session is nil it authenticates first; otherwise it reuses the provided
// session so the browser flow is not repeated.
func switchToAccount(cfg *config.Config, session *aws.Session, account *config.Account) {
	fmt.Printf("Selected: %s (%s) - %s\n", account.Name, account.ID, account.RoleName)

	if session == nil {
		var err error
		if session, err = login(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	// Get credentials
	creds, err := session.GetCredentials(account.ID, account.RoleName)
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
