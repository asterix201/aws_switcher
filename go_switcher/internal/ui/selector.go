package ui

import (
	"fmt"

	"aws-switcher/internal/config"

	"github.com/ktr0731/go-fuzzyfinder"
)

// SelectAccount allows user to select an account using fuzzy finder
func SelectAccount(accounts []config.Account) (*config.Account, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts available")
	}

	idx, err := fuzzyfinder.Find(
		accounts,
		func(i int) string {
			account := accounts[i]
			// Format: AccountName (AccountID) RoleName
			return fmt.Sprintf("%s (%s) %s", account.Name, account.ID, account.RoleName)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			account := accounts[i]
			return fmt.Sprintf("Account: %s\nID: %s\nRole: %s",
				account.Name, account.ID, account.RoleName)
		}),
	)
	if err != nil {
		return nil, err
	}

	return &accounts[idx], nil
}
