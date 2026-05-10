package calendar

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	httpClient    *http.Client // kept for raw requests (e.g. sync-token PROPFIND)
	origin        string       // scheme + host, e.g. "https://apidata.googleusercontent.com"
	principalPath string       // path-only form of the principal URL (starts with "/")
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
	return &caldavConn{
		client:        c,
		httpClient:    httpClient,
		origin:        origin,
		principalPath: u.RequestURI(),
	}, nil
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

// queryWindow fetches all calendar objects (ETags + iCal data) whose VEVENT
// overlaps [start, end) using a calendar-query REPORT.
func (c *caldavConn) queryWindow(ctx context.Context, calPath string, start, end time.Time) ([]caldav.CalendarObject, error) {
	return c.client.QueryCalendar(ctx, calPath, &caldav.CalendarQuery{
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
	})
}

// fetchSyncToken does a Depth:0 PROPFIND to retrieve the current DAV:sync-token
// for a calendar collection. This is called after windowed initial population to
// anchor incremental sync going forward.
func (c *caldavConn) fetchSyncToken(ctx context.Context, calPath string) (string, error) {
	const body = `<?xml version="1.0" encoding="UTF-8"?>` +
		`<D:propfind xmlns:D="DAV:"><D:prop><D:sync-token/></D:prop></D:propfind>`

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", c.origin+calPath, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type xmlProp struct {
		SyncToken string `xml:"sync-token"`
	}
	type xmlPropstat struct {
		Prop   xmlProp `xml:"prop"`
		Status string  `xml:"status"`
	}
	type xmlResponse struct {
		Propstats []xmlPropstat `xml:"propstat"`
	}
	type xmlMultistatus struct {
		Responses []xmlResponse `xml:"response"`
	}

	var ms xmlMultistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return "", fmt.Errorf("parse sync-token response: %w", err)
	}
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if strings.Contains(ps.Status, "200") && ps.Prop.SyncToken != "" {
				return ps.Prop.SyncToken, nil
			}
		}
	}
	return "", fmt.Errorf("sync-token not found in PROPFIND response")
}

// syncCalendar performs a sync-collection REPORT (RFC 6578). It only requests
// ETags so the response stays small. Pass an empty syncToken for the initial
// sync, which returns all objects.
func (c *caldavConn) syncCalendar(ctx context.Context, calPath, syncToken string) (*caldav.SyncResponse, error) {
	return c.client.SyncCollection(ctx, calPath, &caldav.SyncQuery{
		SyncToken:   syncToken,
		CompRequest: caldav.CalendarCompRequest{Name: "VCALENDAR"},
	})
}

const fetchBatchSize = 50

// fetchObjects retrieves full calendar data for a list of object paths via
// calendar-multiget, sending at most fetchBatchSize hrefs per request.
func (c *caldavConn) fetchObjects(ctx context.Context, calPath string, paths []string) ([]caldav.CalendarObject, error) {
	compRequest := caldav.CalendarCompRequest{
		Name:     "VCALENDAR",
		AllProps: true,
		Comps: []caldav.CalendarCompRequest{{
			Name:     "VEVENT",
			AllProps: true,
		}},
	}
	var all []caldav.CalendarObject
	for i := 0; i < len(paths); i += fetchBatchSize {
		end := i + fetchBatchSize
		if end > len(paths) {
			end = len(paths)
		}
		objects, err := c.client.MultiGetCalendar(ctx, calPath, &caldav.CalendarMultiGet{
			Paths:       paths[i:end],
			CompRequest: compRequest,
		})
		if err != nil {
			return nil, err
		}
		all = append(all, objects...)
	}
	return all, nil
}
