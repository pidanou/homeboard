package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pidanou/homeboard/internal/repository"
)

// tickInterval must match the interval CheckAndSend is invoked on (see cmd/server/main.go).
// It defines the window width: an item is "in window" if its due/start time falls in
// [now+offset, now+offset+tickInterval).
const tickInterval = 1 * time.Minute

type ReminderService struct {
	reminders repository.ReminderRepository
	tasks     repository.TaskRepository
	events    repository.EventRepository
	push      *PushService
}

func NewReminderService(reminders repository.ReminderRepository, tasks repository.TaskRepository, events repository.EventRepository, push *PushService) *ReminderService {
	return &ReminderService{reminders: reminders, tasks: tasks, events: events, push: push}
}

// CheckAndSend fires reminder pushes for tasks/events falling within a recipient's
// configured minutes-before window. Idempotent: safe to call every tick.
func (s *ReminderService) CheckAndSend(ctx context.Context) {
	recipients, err := s.reminders.ListRecipients(ctx)
	if err != nil {
		log.Printf("reminder: list recipients: %v", err)
		return
	}

	byFamily := map[string][]*repository.ReminderRecipient{}
	for _, r := range recipients {
		byFamily[r.FamilyID] = append(byFamily[r.FamilyID], r)
	}

	now := time.Now().UTC()
	for familyID, recips := range byFamily {
		s.checkFamily(ctx, familyID, recips, now)
	}
}

func (s *ReminderService) checkFamily(ctx context.Context, familyID string, recips []*repository.ReminderRecipient, now time.Time) {
	tasks, err := s.tasks.ListByFamilyID(ctx, familyID)
	if err != nil {
		log.Printf("reminder: list tasks for family %s: %v", familyID, err)
		return
	}

	maxMinutes := 0
	for _, r := range recips {
		if r.MinutesBefore > maxMinutes {
			maxMinutes = r.MinutesBefore
		}
	}
	events, err := s.events.ListByFamilyAndRange(ctx, familyID, now, now.Add(time.Duration(maxMinutes)*time.Minute+tickInterval))
	if err != nil {
		log.Printf("reminder: list events for family %s: %v", familyID, err)
		return
	}

	for _, r := range recips {
		windowStart := now.Add(time.Duration(r.MinutesBefore) * time.Minute)
		windowEnd := windowStart.Add(tickInterval)

		for _, t := range tasks {
			if t.EndDate == nil || t.Status == "done" || t.AssignedTo == nil || *t.AssignedTo != r.UserID {
				continue
			}
			if inWindow(*t.EndDate, windowStart, windowEnd) {
				s.maybeSend(ctx, r.UserID, familyID, "task", t.ID, *t.EndDate,
					"Task due soon", fmt.Sprintf("%q is due in %d minutes", t.Title, r.MinutesBefore))
			}
		}

		for _, e := range events {
			if e.Cancelled {
				continue
			}
			if len(e.AttendeeIDs) > 0 && !contains(e.AttendeeIDs, r.UserID) {
				continue
			}
			if inWindow(e.StartAt, windowStart, windowEnd) {
				s.maybeSend(ctx, r.UserID, familyID, "event", e.ID, e.StartAt,
					"Event starting soon", fmt.Sprintf("%q starts in %d minutes", e.Title, r.MinutesBefore))
			}
		}
	}
}

func (s *ReminderService) maybeSend(ctx context.Context, userID, familyID, itemType, itemID string, occurrenceAt time.Time, title, body string) {
	sent, err := s.reminders.HasSent(ctx, userID, itemType, itemID, occurrenceAt)
	if err != nil {
		log.Printf("reminder: check sent: %v", err)
		return
	}
	if sent {
		return
	}

	url := fmt.Sprintf("/households/%s/board", familyID)
	if itemType == "event" {
		url = fmt.Sprintf("/households/%s/calendar", familyID)
	}
	s.push.SendToUser(ctx, userID, title, body, url)

	if err := s.reminders.MarkSent(ctx, userID, itemType, itemID, occurrenceAt); err != nil {
		log.Printf("reminder: mark sent: %v", err)
	}
}

func inWindow(t, start, end time.Time) bool {
	return !t.Before(start) && t.Before(end)
}

func contains(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
