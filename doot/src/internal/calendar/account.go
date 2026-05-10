package calendar

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Account holds non-secret CalDAV account info (token lives in keyring).
type Account struct {
	Name     string `json:"name"`     // Display name, e.g. "Work" or "Personal"
	URL      string `json:"url"`      // CalDAV principal or home-set URL
	Username string `json:"username"` // Usually the account email
}

func accountsDir() string {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(dir, "doot", "caldav")
}

func accountPath(username string) string {
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(username)
	return filepath.Join(accountsDir(), safe+".json")
}

// SaveAccount persists a CalDAV account to disk (mode 0600).
func SaveAccount(acc *Account) error {
	if err := os.MkdirAll(accountsDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(acc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(accountPath(acc.Username), data, 0600)
}

// LoadAccount reads a stored account by username.
func LoadAccount(username string) (*Account, error) {
	data, err := os.ReadFile(accountPath(username))
	if err != nil {
		return nil, err
	}
	var acc Account
	return &acc, json.Unmarshal(data, &acc)
}

// ListAccounts returns all stored CalDAV accounts.
func ListAccounts() ([]*Account, error) {
	entries, err := os.ReadDir(accountsDir())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var accounts []*Account
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) != ".json" || strings.HasSuffix(name, ".sync.json") {
			continue
		}
		username := strings.TrimSuffix(name, ".json")
		acc, err := LoadAccount(username)
		if err != nil {
			continue
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

// RemoveAccount deletes a stored account.
func RemoveAccount(username string) error {
	path := accountPath(username)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %s: %w", username, err)
	}
	return nil
}

// GoogleCalDAVURL returns the standard Google CalDAV URL for a Gmail address.
func GoogleCalDAVURL(email string) string {
	return "https://apidata.googleusercontent.com/caldav/v2/" + email + "/user"
}
