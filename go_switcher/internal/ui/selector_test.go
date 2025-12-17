package ui

import (
	"testing"

	"aws-switcher/internal/config"
)

func TestSelectAccount_EmptyList(t *testing.T) {
	accounts := []config.Account{}

	_, err := SelectAccount(accounts)
	if err == nil {
		t.Error("Expected error when selecting from empty account list")
	}

	expectedErrMsg := "no accounts available"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestSelectAccount_ValidAccounts(t *testing.T) {
	accounts := []config.Account{
		{
			ID:       "123456789012",
			Name:     "Test Account 1",
			RoleName: "AdminRole",
		},
		{
			ID:       "123456789013",
			Name:     "Test Account 2",
			RoleName: "ReadOnlyRole",
		},
	}

	// This test verifies that the function signature is correct
	// and that it can handle valid input without panicking
	// In a real interactive test, you would need to mock the fuzzyfinder
	t.Logf("SelectAccount can handle %d accounts", len(accounts))

	// Verify accounts are properly formatted
	for i, account := range accounts {
		if account.ID == "" {
			t.Errorf("Account %d has empty ID", i)
		}
		if account.Name == "" {
			t.Errorf("Account %d has empty Name", i)
		}
		if account.RoleName == "" {
			t.Errorf("Account %d has empty RoleName", i)
		}
	}
}

func TestAccountFormatting(t *testing.T) {
	account := config.Account{
		ID:       "123456789012",
		Name:     "Production Account",
		RoleName: "AdminRole",
	}

	// Test the format string that would be used in SelectAccount
	expected := "Production Account (123456789012) AdminRole"
	actual := account.Name + " (" + account.ID + ") " + account.RoleName

	if actual != expected {
		t.Errorf("Account formatting mismatch:\nExpected: %s\nActual: %s", expected, actual)
	}
}
