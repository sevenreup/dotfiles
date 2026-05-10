package calendar

import (
	"context"
	"fmt"
	"time"
)

// Event is the wire format for a calendar event sent to QML.
type Event struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Start        string `json:"start"`        // "YYYY-MM-DD" (all-day) or RFC3339 (timed)
	End          string `json:"end"`
	AllDay       bool   `json:"allDay"`
	Color        string `json:"color"`
	AccountEmail string `json:"accountEmail"` // actually the CalDAV username
	CalendarName string `json:"calendarName"`
}

// syncWindowDays is the half-width of the initial sync window on each side of today.
const syncWindowDays = 180

// FetchMonth syncs all accounts and returns events for year/month from the
// local cache.
func FetchMonth(year, month int) ([]*Event, error) {
	accounts, err := ListAccounts()
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, nil
	}

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	var all []*Event
	for _, acc := range accounts {
		evs, err := syncAndFetchForAccount(acc, start, end)
		if err != nil {
			fmt.Printf("[calendar] %s: %v\n", acc.Username, err)
			continue
		}
		all = append(all, evs...)
	}
	return all, nil
}

func syncAndFetchForAccount(acc *Account, start, end time.Time) ([]*Event, error) {
	httpClient, err := newBearerClient(acc.Username)
	if err != nil {
		fmt.Printf("[calendar] %s: token missing, removing account\n", acc.Username)
		RemoveAccount(acc.Username)
		return nil, fmt.Errorf("auth: %w", err)
	}

	ctx := context.Background()

	conn, err := newCaldavConn(httpClient, acc.URL)
	if err != nil {
		return nil, err
	}

	state, err := loadSyncState(acc.Username)
	if err != nil {
		return nil, fmt.Errorf("load sync state: %w", err)
	}

	homeSet, err := conn.discoverHomeSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("home-set discovery: %w", err)
	}

	cals, err := conn.listCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("list calendars: %w", err)
	}

	stateChanged := false
	for _, cal := range cals {
		calState, ok := state.Calendars[cal.Path]
		if !ok {
			calState = &calendarSyncState{
				Name:    cal.Name,
				Color:   cal.Color,
				Objects: make(map[string]string),
			}
			state.Calendars[cal.Path] = calState
		}

		var changed bool
		if calState.SyncToken == "" {
			fmt.Printf("[calendar] %s → %s: initial sync (window ±%dd)\n", acc.Username, cal.Name, syncWindowDays)
			changed, err = doInitialSync(ctx, conn, cal.Path, calState)
		} else {
			fmt.Printf("[calendar] %s → %s: incremental sync\n", acc.Username, cal.Name)
			changed, err = doIncrementalSync(ctx, conn, cal.Path, calState)
		}
		if err != nil {
			fmt.Printf("[calendar] %s → %s: %v\n", acc.Username, cal.Name, err)
			continue
		}
		if changed {
			stateChanged = true
		}
	}

	if stateChanged {
		if err := saveSyncState(acc.Username, state); err != nil {
			fmt.Printf("[calendar] %s: save sync state: %v\n", acc.Username, err)
		}
	}

	// Parse all cached objects, filter to the requested time range.
	var events []*Event
	for _, calState := range state.Calendars {
		for _, raw := range calState.Objects {
			cal, err := decodeIcal(raw)
			if err != nil {
				continue
			}
			evs, _ := parseEvents(cal, acc.Username, calState.Name, calState.Color)
			for _, ev := range evs {
				if eventInRange(ev, start, end) {
					events = append(events, ev)
				}
			}
		}
	}
	return events, nil
}

// doInitialSync populates calState using two bounded calendar-query windows
// (past and future), then fetches the current sync-token to anchor future
// incremental syncs. Path is used as the dedup key so events that overlap
// both windows are stored once.
func doInitialSync(ctx context.Context, conn *caldavConn, calPath string, calState *calendarSyncState) (bool, error) {
	now := time.Now().UTC()
	windows := [2][2]time.Time{
		{now.AddDate(0, 0, -syncWindowDays), now},
		{now, now.AddDate(0, 0, syncWindowDays)},
	}

	for _, w := range windows {
		label := w[0].Format("2006-01-02") + " → " + w[1].Format("2006-01-02")
		fmt.Printf("[calendar]   window %s: fetching…\n", label)
		objects, err := conn.queryWindow(ctx, calPath, w[0], w[1])
		if err != nil {
			return false, fmt.Errorf("query window %s: %w", label, err)
		}
		for _, obj := range objects {
			if obj.Data == nil {
				continue
			}
			raw, err := encodeIcal(obj.Data)
			if err == nil {
				calState.Objects[obj.Path] = raw
			}
		}
		fmt.Printf("[calendar]   window %s: %d objects\n", label, len(objects))
	}

	fmt.Printf("[calendar]   total cached: %d objects — fetching sync-token\n", len(calState.Objects))

	// Anchor incremental sync to the current state of the calendar.
	token, err := conn.fetchSyncToken(ctx, calPath)
	if err != nil {
		return false, fmt.Errorf("fetch sync-token: %w", err)
	}
	calState.SyncToken = token
	fmt.Printf("[calendar]   sync-token acquired, initial sync complete\n")

	return true, nil
}

// doIncrementalSync uses sync-collection to fetch only objects changed since
// the last stored sync-token, then multigets their data in batches.
func doIncrementalSync(ctx context.Context, conn *caldavConn, calPath string, calState *calendarSyncState) (bool, error) {
	syncResp, err := conn.syncCalendar(ctx, calPath, calState.SyncToken)
	if err != nil {
		return false, fmt.Errorf("sync-collection: %w", err)
	}

	if len(syncResp.Updated) == 0 && len(syncResp.Deleted) == 0 {
		fmt.Printf("[calendar]   no changes\n")
		return false, nil
	}

	fmt.Printf("[calendar]   %d updated, %d deleted\n", len(syncResp.Updated), len(syncResp.Deleted))

	if len(syncResp.Updated) > 0 {
		paths := make([]string, len(syncResp.Updated))
		for i, obj := range syncResp.Updated {
			paths[i] = obj.Path
		}
		batches := (len(paths) + fetchBatchSize - 1) / fetchBatchSize
		fmt.Printf("[calendar]   fetching %d objects in %d batch(es)\n", len(paths), batches)
		objects, err := conn.fetchObjects(ctx, calPath, paths)
		if err != nil {
			return false, fmt.Errorf("multiget: %w", err)
		}
		for _, obj := range objects {
			if obj.Data == nil {
				continue
			}
			raw, err := encodeIcal(obj.Data)
			if err == nil {
				calState.Objects[obj.Path] = raw
			}
		}
	}

	for _, path := range syncResp.Deleted {
		delete(calState.Objects, path)
	}

	calState.SyncToken = syncResp.SyncToken
	fmt.Printf("[calendar]   done, %d objects in cache\n", len(calState.Objects))
	return true, nil
}
