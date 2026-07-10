package service

import (
	"time"

	ics "github.com/arran4/golang-ical"
)

func propValue(vevent *ics.VEvent, prop ics.ComponentProperty) string {
	p := vevent.GetProperty(prop)
	if p == nil {
		return ""
	}
	return p.Value
}

func isAllDay(vevent *ics.VEvent) bool {
	p := vevent.GetProperty(ics.ComponentPropertyDtStart)
	if p == nil {
		return false
	}
	values := p.ICalParameters["VALUE"]
	return len(values) > 0 && values[0] == "DATE"
}

// eventTimes extracts start/end from a VEVENT, defaulting a missing/invalid DTEND
// to a sane duration after DTSTART (external feeds don't always include one).
func eventTimes(vevent *ics.VEvent, allDay bool) (start, end time.Time, ok bool) {
	var err error
	if allDay {
		start, err = vevent.GetAllDayStartAt()
	} else {
		start, err = vevent.GetStartAt()
	}
	if err != nil {
		return time.Time{}, time.Time{}, false
	}

	if allDay {
		end, err = vevent.GetAllDayEndAt()
	} else {
		end, err = vevent.GetEndAt()
	}
	if err != nil || !end.After(start) {
		if allDay {
			end = start.AddDate(0, 0, 1)
		} else {
			end = start.Add(time.Hour)
		}
	}
	return start.UTC(), end.UTC(), true
}
