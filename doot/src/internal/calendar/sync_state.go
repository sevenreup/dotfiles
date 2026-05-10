package calendar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// calendarSyncState holds the sync token and cached raw iCal objects for one
// calendar collection.
type calendarSyncState struct {
	Name      string            `json:"name"`
	Color     string            `json:"color"`
	SyncToken string            `json:"syncToken"`
	Objects   map[string]string `json:"objects"` // object path → raw iCal
}

// accountSyncState is the top-level structure persisted per account.
type accountSyncState struct {
	Calendars map[string]*calendarSyncState `json:"calendars"` // calendar path → state
}

func syncStatePath(username string) string {
	safe := strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(username)
	return filepath.Join(accountsDir(), safe+".sync.json")
}

func loadSyncState(username string) (*accountSyncState, error) {
	data, err := os.ReadFile(syncStatePath(username))
	if os.IsNotExist(err) {
		return &accountSyncState{Calendars: make(map[string]*calendarSyncState)}, nil
	}
	if err != nil {
		return nil, err
	}
	var state accountSyncState
	if err := json.Unmarshal(data, &state); err != nil {
		// Corrupt state: start fresh rather than hard-failing.
		return &accountSyncState{Calendars: make(map[string]*calendarSyncState)}, nil
	}
	if state.Calendars == nil {
		state.Calendars = make(map[string]*calendarSyncState)
	}
	return &state, nil
}

func saveSyncState(username string, state *accountSyncState) error {
	if err := os.MkdirAll(accountsDir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(syncStatePath(username), data, 0600)
}
