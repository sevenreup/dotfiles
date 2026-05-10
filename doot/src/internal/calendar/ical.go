package calendar

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	ical "github.com/emersion/go-ical"
)

// parseEvents extracts Event objects from a decoded iCalendar object.
// calColor and calName are applied as defaults when the event has no color.
func parseEvents(cal *ical.Calendar, accountUsername, calName, calColor string) ([]*Event, error) {
	var events []*Event
	for _, comp := range cal.Children {
		if comp.Name != ical.CompEvent {
			continue
		}

		ev, err := componentToEvent(comp, accountUsername, calName, calColor)
		if err != nil {
			continue // skip malformed events
		}
		events = append(events, ev)
	}
	return events, nil
}

func componentToEvent(comp *ical.Component, accountUsername, calName, calColor string) (*Event, error) {
	uid := propStr(comp, ical.PropUID)
	summary := propStr(comp, ical.PropSummary)
	if summary == "" {
		summary = "(no title)"
	}
	status := strings.ToUpper(propStr(comp, ical.PropStatus))
	if status == "CANCELLED" {
		return nil, fmt.Errorf("cancelled")
	}

	// Resolve color: event color → calendar color
	color := calColor
	if color == "" {
		color = "#cba6f7"
	}

	// Determine start/end and all-day status
	dtstart := comp.Props.Get(ical.PropDateTimeStart)
	dtend := comp.Props.Get(ical.PropDateTimeEnd)

	if dtstart == nil {
		return nil, fmt.Errorf("event %q has no DTSTART", uid)
	}

	allDay := isDateOnly(dtstart)
	var startStr, endStr string

	if allDay {
		// go-ical uses DateTime() for both DATE and DATE-TIME values.
		// For DATE, it returns midnight in the given location.
		t, err := dtstart.DateTime(time.Local)
		if err != nil {
			return nil, err
		}
		startStr = t.Format("2006-01-02")

		if dtend != nil {
			te, err := dtend.DateTime(time.Local)
			if err == nil {
				endStr = te.Format("2006-01-02")
			}
		}
		if endStr == "" {
			endStr = startStr
		}
	} else {
		t, err := dtstart.DateTime(time.Local)
		if err != nil {
			return nil, err
		}
		startStr = t.Format(time.RFC3339)

		if dtend != nil {
			te, err := dtend.DateTime(time.Local)
			if err == nil {
				endStr = te.Format(time.RFC3339)
			}
		}
		if endStr == "" {
			endStr = startStr
		}
	}

	return &Event{
		ID:           uid,
		Title:        summary,
		Start:        startStr,
		End:          endStr,
		AllDay:       allDay,
		Color:        color,
		AccountEmail: accountUsername,
		CalendarName: calName,
	}, nil
}

// isDateOnly returns true when the DTSTART is a DATE (not DATE-TIME).
func isDateOnly(prop *ical.Prop) bool {
	vtype := prop.Params.Get("VALUE")
	if strings.EqualFold(vtype, "DATE") {
		return true
	}
	// No time component and no 'T' in the value → DATE
	return !strings.Contains(prop.Value, "T")
}

func propStr(comp *ical.Component, name string) string {
	p := comp.Props.Get(name)
	if p == nil {
		return ""
	}
	return p.Value
}

// encodeIcal serialises a decoded *ical.Calendar back to its text form for
// caching. Re-parsing with ical.NewDecoder is lossless for the properties we
// care about.
func encodeIcal(cal *ical.Calendar) (string, error) {
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// decodeIcal is the inverse of encodeIcal.
func decodeIcal(raw string) (*ical.Calendar, error) {
	return ical.NewDecoder(strings.NewReader(raw)).Decode()
}

// eventInRange reports whether ev starts within [start, end).
func eventInRange(ev *Event, start, end time.Time) bool {
	var t time.Time
	var err error
	if ev.AllDay {
		t, err = time.ParseInLocation("2006-01-02", ev.Start, time.Local)
	} else {
		t, err = time.Parse(time.RFC3339, ev.Start)
	}
	if err != nil {
		return false
	}
	return !t.Before(start) && t.Before(end)
}
