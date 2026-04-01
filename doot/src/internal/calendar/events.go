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

// FetchMonth returns all events for year/month across all stored CalDAV accounts.
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
		evs, err := fetchForAccount(acc, start, end)
		if err != nil {
			fmt.Printf("[calendar] %s: %v\n", acc.Username, err)
			continue
		}
		all = append(all, evs...)
	}
	return all, nil
}

func fetchForAccount(acc *Account, start, end time.Time) ([]*Event, error) {
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

	homeSet, err := conn.discoverHomeSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("home-set discovery: %w", err)
	}

	cals, err := conn.listCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("list calendars: %w", err)
	}

	var events []*Event
	for _, cal := range cals {
		icals, err := conn.queryEvents(ctx, cal.Path, start, end)
		if err != nil {
			fmt.Printf("[calendar] %s → %s: %v\n", acc.Username, cal.Name, err)
			continue
		}
		for _, icalCal := range icals {
			evs, err := parseEvents(icalCal, acc.Username, cal.Name, cal.Color)
			if err != nil {
				continue
			}
			events = append(events, evs...)
		}
	}
	return events, nil
}
