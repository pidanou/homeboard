package service

import (
	"bytes"
	"context"
	"fmt"

	ics "github.com/arran4/golang-ical"
)

type CalendarImportService struct {
	events *EventService
}

func NewCalendarImportService(events *EventService) *CalendarImportService {
	return &CalendarImportService{events: events}
}

// ImportFile parses an uploaded .ics file and creates a plain household event for each
// valid VEVENT. Events missing a title or a usable start time are skipped, not failed.
func (s *CalendarImportService) ImportFile(ctx context.Context, familyID, userID string, data []byte) (imported, skipped int, err error) {
	cal, err := ics.ParseCalendar(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("parse ics: %w", err)
	}

	for _, vevent := range cal.Events() {
		title := propValue(vevent, ics.ComponentPropertySummary)
		allDay := isAllDay(vevent)
		start, end, ok := eventTimes(vevent, allDay)
		if title == "" || !ok {
			skipped++
			continue
		}

		var rrule *string
		if r := propValue(vevent, ics.ComponentPropertyRrule); r != "" {
			rrule = &r
		}

		_, err := s.events.Create(ctx, familyID, userID,
			title,
			propValue(vevent, ics.ComponentPropertyDescription),
			propValue(vevent, ics.ComponentPropertyLocation),
			start, end, allDay, nil, nil, rrule, "default", nil, false,
		)
		if err != nil {
			skipped++
			continue
		}
		imported++
	}

	return imported, skipped, nil
}
