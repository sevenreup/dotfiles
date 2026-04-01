package calendar

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CalendarInfo is a calendar returned from PROPFIND.
type CalendarInfo struct {
	Path  string
	Name  string
	Color string
}

// newHTTPClient returns an http.Client with Basic Auth transport.
func newHTTPClient(username, password string) *http.Client {
	return &http.Client{
		Transport: &basicAuthTransport{
			username: username,
			password: password,
			inner:    http.DefaultTransport,
		},
	}
}

type basicAuthTransport struct {
	username, password string
	inner              http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.SetBasicAuth(t.username, t.password)
	return t.inner.RoundTrip(r2)
}

// ──────────────────────────────────────────────────────────────────────────────
// PROPFIND: discover calendar home-set
// ──────────────────────────────────────────────────────────────────────────────

func discoverHomeSet(client *http.Client, principalURL string) (string, error) {
	const body = `<?xml version="1.0" encoding="UTF-8"?>` +
		`<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">` +
		`<D:prop><C:calendar-home-set/></D:prop>` +
		`</D:propfind>`

	resp, err := propfind(client, principalURL, "0", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	type xmlHref struct {
		Value string `xml:",chardata"`
	}
	type xmlHomeSet struct {
		Hrefs []xmlHref `xml:"href"`
	}
	type xmlProp struct {
		HomeSet *xmlHomeSet `xml:"calendar-home-set"`
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
		return "", fmt.Errorf("parse home-set response: %w", err)
	}

	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if !strings.Contains(ps.Status, "200") {
				continue
			}
			if ps.Prop.HomeSet != nil && len(ps.Prop.HomeSet.Hrefs) > 0 {
				href := ps.Prop.HomeSet.Hrefs[0].Value
				return resolveURL(principalURL, href), nil
			}
		}
	}
	// Fall back: treat the principal URL's parent as the home set
	return principalURL, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// PROPFIND: list calendars
// ──────────────────────────────────────────────────────────────────────────────

func listCalendars(client *http.Client, homeSetURL string) ([]CalendarInfo, error) {
	const body = `<?xml version="1.0" encoding="UTF-8"?>` +
		`<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:A="http://apple.com/ns/ical/">` +
		`<D:prop><D:resourcetype/><D:displayname/><A:calendar-color/></D:prop>` +
		`</D:propfind>`

	resp, err := propfind(client, homeSetURL, "1", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type xmlCalendar struct {
		XMLName xml.Name `xml:"calendar"`
	}
	type xmlResourceType struct {
		Calendar *xmlCalendar `xml:"calendar"`
	}
	type xmlProp struct {
		ResourceType  xmlResourceType `xml:"resourcetype"`
		DisplayName   string          `xml:"displayname"`
		CalendarColor string          `xml:"calendar-color"`
	}
	type xmlPropstat struct {
		Prop   xmlProp `xml:"prop"`
		Status string  `xml:"status"`
	}
	type xmlResponse struct {
		Href      string        `xml:"href"`
		Propstats []xmlPropstat `xml:"propstat"`
	}
	type xmlMultistatus struct {
		Responses []xmlResponse `xml:"response"`
	}

	var ms xmlMultistatus
	if err := xml.NewDecoder(resp.Body).Decode(&ms); err != nil {
		return nil, fmt.Errorf("parse calendar list: %w", err)
	}

	var calendars []CalendarInfo
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if !strings.Contains(ps.Status, "200") {
				continue
			}
			if ps.Prop.ResourceType.Calendar == nil {
				continue // not a calendar resource
			}
			color := strings.TrimSpace(ps.Prop.CalendarColor)
			if color == "" {
				color = "#cba6f7"
			}
			// Strip alpha if present (e.g. "#3a7bd5ff" → "#3a7bd5")
			if len(color) == 9 && color[0] == '#' {
				color = color[:7]
			}
			calendars = append(calendars, CalendarInfo{
				Path:  resolveURL(homeSetURL, r.Href),
				Name:  ps.Prop.DisplayName,
				Color: color,
			})
		}
	}
	return calendars, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// REPORT: fetch events in time range
// ──────────────────────────────────────────────────────────────────────────────

func queryEvents(client *http.Client, calendarURL string, start, end time.Time) ([]string, error) {
	tFmt := "20060102T150405Z"
	body := `<?xml version="1.0" encoding="UTF-8"?>` +
		`<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">` +
		`<D:prop><C:calendar-data/></D:prop>` +
		`<C:filter>` +
		`<C:comp-filter name="VCALENDAR">` +
		`<C:comp-filter name="VEVENT">` +
		`<C:time-range start="` + start.UTC().Format(tFmt) + `" end="` + end.UTC().Format(tFmt) + `"/>` +
		`</C:comp-filter>` +
		`</C:comp-filter>` +
		`</C:filter>` +
		`</C:calendar-query>`

	req, err := http.NewRequest("REPORT", calendarURL, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("REPORT %s: %w", calendarURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("REPORT %s: HTTP %d", calendarURL, resp.StatusCode)
	}

	type xmlProp struct {
		CalendarData string `xml:"calendar-data"`
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
		return nil, fmt.Errorf("parse REPORT response: %w", err)
	}

	var icals []string
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if !strings.Contains(ps.Status, "200") {
				continue
			}
			if data := strings.TrimSpace(ps.Prop.CalendarData); data != "" {
				icals = append(icals, data)
			}
		}
	}
	return icals, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func propfind(client *http.Client, url, depth, body string) (*http.Response, error) {
	req, err := http.NewRequest("PROPFIND", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", depth)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PROPFIND %s: %w", url, err)
	}
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusMultiStatus {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("PROPFIND %s: HTTP %d", url, resp.StatusCode)
	}
	return resp, nil
}

// resolveURL combines a base URL with an href that may be absolute or path-only.
func resolveURL(base, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	// href is a path — extract scheme+host from base
	parts := strings.SplitN(base, "/", 4)
	if len(parts) >= 3 {
		return parts[0] + "//" + parts[2] + href
	}
	return base
}
