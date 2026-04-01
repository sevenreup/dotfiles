package calendar

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	ical "github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

// CalendarInfo holds calendar metadata returned from the server.
type CalendarInfo struct {
	Path  string
	Name  string
	Color string
}

// caldavConn wraps a go-webdav CalDAV client for a single account.
type caldavConn struct {
	client        *caldav.Client
	principalPath string // path-only form of the principal URL (starts with "/")
}

// newCaldavConn creates a CalDAV client. go-webdav's ResolveHref treats paths
// starting with "/" as absolute on the server, so we split the principal URL
// into origin (used as the client endpoint) and path.
func newCaldavConn(httpClient *http.Client, principalURL string) (*caldavConn, error) {
	u, err := url.Parse(principalURL)
	if err != nil {
		return nil, fmt.Errorf("parse principal URL: %w", err)
	}
	origin := u.Scheme + "://" + u.Host
	c, err := caldav.NewClient(httpClient, origin)
	if err != nil {
		return nil, fmt.Errorf("caldav client: %w", err)
	}
	return &caldavConn{client: c, principalPath: u.RequestURI()}, nil
}

func (c *caldavConn) discoverHomeSet(ctx context.Context) (string, error) {
	return c.client.FindCalendarHomeSet(ctx, c.principalPath)
}

func (c *caldavConn) listCalendars(ctx context.Context, homeSet string) ([]CalendarInfo, error) {
	cals, err := c.client.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, err
	}
	result := make([]CalendarInfo, 0, len(cals))
	for _, cal := range cals {
		result = append(result, CalendarInfo{
			Path:  cal.Path,
			Name:  cal.Name,
			Color: "#cba6f7",
		})
	}
	return result, nil
}

func (c *caldavConn) queryEvents(ctx context.Context, calPath string, start, end time.Time) ([]*ical.Calendar, error) {
	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:     "VCALENDAR",
			AllProps: true,
			Comps: []caldav.CalendarCompRequest{{
				Name:     "VEVENT",
				AllProps: true,
			}},
		},
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VEVENT",
				Start: start,
				End:   end,
			}},
		},
	}
	objects, err := c.client.QueryCalendar(ctx, calPath, query)
	if err != nil {
		return nil, err
	}
	cals := make([]*ical.Calendar, 0, len(objects))
	for i := range objects {
		if objects[i].Data != nil {
			cals = append(cals, objects[i].Data)
		}
	}
	return cals, nil
}
