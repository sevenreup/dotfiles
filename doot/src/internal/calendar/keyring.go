package calendar

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	keychainService = "doot"
	credentialsKey  = "google-oauth-credentials"
	tokenKeyPrefix  = "google-oauth-token-"
)

type storedCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// StoreCredentials saves Google OAuth2 client credentials to the GNOME Keyring.
func StoreCredentials(clientID, clientSecret string) error {
	data, err := json.Marshal(storedCredentials{ClientID: clientID, ClientSecret: clientSecret})
	if err != nil {
		return err
	}
	return keyring.Set(keychainService, credentialsKey, string(data))
}

// LoadCredentialValues reads Google OAuth2 client credentials from the keyring.
func LoadCredentialValues() (clientID, clientSecret string, err error) {
	raw, err := keyring.Get(keychainService, credentialsKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", "", fmt.Errorf(
				"Google OAuth credentials not in keyring\n" +
					"Run: doot calendar setup credentials",
			)
		}
		return "", "", fmt.Errorf("keyring error: %w", err)
	}
	var creds storedCredentials
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return "", "", fmt.Errorf("invalid credentials in keyring: %w", err)
	}
	return creds.ClientID, creds.ClientSecret, nil
}

// HasCredentials returns true if Google OAuth2 credentials are in the keyring.
func HasCredentials() bool {
	_, err := keyring.Get(keychainService, credentialsKey)
	return err == nil
}

// StoreToken persists an OAuth2 token for the given email in the keyring.
func StoreToken(email string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return keyring.Set(keychainService, tokenKeyPrefix+email, string(data))
}

// LoadToken reads a stored OAuth2 token for the given email.
func LoadToken(email string) (*oauth2.Token, error) {
	raw, err := keyring.Get(keychainService, tokenKeyPrefix+email)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, fmt.Errorf("no token for %s — run: doot calendar auth add", email)
		}
		return nil, fmt.Errorf("keyring error for %s: %w", email, err)
	}
	var token oauth2.Token
	return &token, json.Unmarshal([]byte(raw), &token)
}

// DeleteToken removes an OAuth2 token from the keyring.
func DeleteToken(email string) error {
	if err := keyring.Delete(keychainService, tokenKeyPrefix+email); err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("keyring delete for %s: %w", email, err)
	}
	return nil
}
