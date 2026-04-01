package calendar

// Manager coordinates calendar operations for the doot server.
type Manager struct{}

func New() *Manager { return &Manager{} }

// Accounts returns all connected accounts.
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

// RemoveAccount deletes a stored account and its keyring token.
func (m *Manager) RemoveAccount(username string) error {
	_ = DeleteToken(username) // best-effort
	return RemoveAccount(username)
}
