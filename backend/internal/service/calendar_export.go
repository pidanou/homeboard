package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/pidanou/homeboard/internal/model"
	"github.com/pidanou/homeboard/internal/repository"
)

type CalendarExportService struct {
	tokens     repository.CalendarExportTokenRepository
	events     repository.EventRepository
	households repository.HouseholdRepository
}

func NewCalendarExportService(tokens repository.CalendarExportTokenRepository, events repository.EventRepository, households repository.HouseholdRepository) *CalendarExportService {
	return &CalendarExportService{tokens: tokens, events: events, households: households}
}

func (s *CalendarExportService) GetOrNil(ctx context.Context, familyID string) *model.CalendarExportToken {
	t, err := s.tokens.GetByFamilyID(ctx, familyID)
	if err != nil {
		return nil
	}
	return t
}

// Regenerate replaces any existing export token for the family with a new one,
// same delete-then-create shape as InviteService.Create.
func (s *CalendarExportService) Regenerate(ctx context.Context, familyID, userID string) (*model.CalendarExportToken, error) {
	if err := s.tokens.DeleteByFamilyID(ctx, familyID); err != nil {
		return nil, fmt.Errorf("clear existing export token: %w", err)
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	token := &model.CalendarExportToken{
		Token:     hex.EncodeToString(b),
		FamilyID:  familyID,
		CreatedBy: userID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.tokens.Create(ctx, token); err != nil {
		return nil, fmt.Errorf("create export token: %w", err)
	}
	return token, nil
}

func (s *CalendarExportService) Revoke(ctx context.Context, familyID string) error {
	return s.tokens.DeleteByFamilyID(ctx, familyID)
}

// BuildICS renders every non-synced household event as an ICS feed for the given export token.
func (s *CalendarExportService) BuildICS(ctx context.Context, token string) ([]byte, error) {
	t, err := s.tokens.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("export token not found")
	}

	events, err := s.events.ListAllForExport(ctx, t.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("list events for export: %w", err)
	}

	household, err := s.households.GetByID(ctx, t.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("load household: %w", err)
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetXWRCalName(household.Name)

	// Group exceptions by parent event ID so they can be attached as EXDATE/overrides.
	exceptionsByParent := map[string][]*model.Event{}
	var parents []*model.Event
	for _, e := range events {
		if e.RecurrenceParentID != nil {
			exceptionsByParent[*e.RecurrenceParentID] = append(exceptionsByParent[*e.RecurrenceParentID], e)
			continue
		}
		parents = append(parents, e)
	}

	for _, e := range parents {
		vevent := cal.AddEvent(e.ID)
		writeEventFields(vevent, e)
		if e.RecurrenceRule != nil && *e.RecurrenceRule != "" {
			vevent.AddRrule(strings.TrimPrefix(*e.RecurrenceRule, "RRULE:"))
		}
		for _, exc := range exceptionsByParent[e.ID] {
			if exc.Cancelled {
				vevent.AddExdate(icsDateTime(*exc.RecurrenceDate, e.AllDay))
				continue
			}
			override := ics.NewEvent(e.ID)
			writeEventFields(override, exc)
			override.SetProperty(ics.ComponentPropertyRecurrenceId, icsDateTime(*exc.RecurrenceDate, e.AllDay))
			cal.AddVEvent(override)
		}
	}

	return []byte(cal.Serialize()), nil
}

func writeEventFields(vevent *ics.VEvent, e *model.Event) {
	vevent.SetSummary(e.Title)
	if e.Description != "" {
		vevent.SetDescription(e.Description)
	}
	if e.Location != "" {
		vevent.SetLocation(e.Location)
	}
	if e.AllDay {
		vevent.SetAllDayStartAt(e.StartAt)
		vevent.SetAllDayEndAt(e.EndAt)
	} else {
		vevent.SetStartAt(e.StartAt)
		vevent.SetEndAt(e.EndAt)
	}
}

func icsDateTime(t time.Time, allDay bool) string {
	if allDay {
		return t.UTC().Format("20060102")
	}
	return t.UTC().Format("20060102T150405Z")
}
