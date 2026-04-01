package calendar

// Manager coordinates calendar operations for the doot server.
type Manager struct{}

func New() *Manager { return &Manager{} }

// Accounts returns all connected accounts (without passwords).
func (m *Manager) Accounts() ([]map[string]string, error) {
	accounts, err := ListAccounts()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]string, len(accounts))
	for i, a := range accounts {
		out[i] = map[string]string{
			"name":     a.Name,
			"username": a.Username,
			"url":      a.URL,
		}
	}
	return out, nil
}

// Events fetches events for the given year/month across all accounts.
func (m *Manager) Events(year, month int) ([]*Event, error) {
	return FetchMonth(year, month)
}

// AddAccount verifies CalDAV credentials and saves the account.
func (m *Manager) AddAccount(name, url, username, password string) error {
	acc := &Account{
		Name:     name,
		URL:      url,
		Username: username,
		Password: password,
	}
	if _, err := VerifyAccount(acc); err != nil {
		return err
	}
	return SaveAccount(acc)
}

// RemoveAccount deletes a stored account.
func (m *Manager) RemoveAccount(username string) error {
	return RemoveAccount(username)
}
