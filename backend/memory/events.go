package memory

import (
	"errors"
	"log"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/util"
)

type Events struct {
	events map[string][]*event
}

type event struct {
	*backend.Event

	listeners []chan *event
}

// Merge two events.
func mergeEvents(dst, src *backend.Event) *backend.Event {
	if dst == nil {
		dst = &backend.Event{}
	}

	dst.ID = src.ID

	if src.Refresh != 0 {
		dst.Refresh = src.Refresh
	}
	if src.Reload != 0 {
		dst.Reload = src.Reload
	}
	if src.User != nil {
		dst.User = src.User
	}

	dst.Notices = append(dst.Notices, src.Notices...)

	dst.Messages = append(dst.Messages, src.Messages...)
	dst.Conversations = append(dst.Conversations, src.Conversations...)
	dst.Labels = append(dst.Labels, src.Labels...)
	dst.Contacts = append(dst.Contacts, src.Contacts...)

	if src.MessageCounts != nil {
		dst.MessageCounts = src.MessageCounts
	}
	if src.ConversationCounts != nil {
		dst.ConversationCounts = src.ConversationCounts
	}

	return dst
}

func (b *Events) insertEvent(user string, e *backend.Event) error {
	e.ID = util.GenerateId()
	b.events[user] = append(b.events[user], &event{Event: e})
	return nil
}

func (b *Events) InsertEvent(user string, e *backend.Event) error {
	// If there is no listener, do not insert the event
	insert := false
	for _, e := range b.events[user] {
		if len(e.listeners) > 0 {
			insert = true
			break
		}
	}

	log.Println("insert_event", e, insert)

	if insert {
		b.insertEvent(user, e)
	}

	return nil
}

// Listen for an event.
func (b *Events) listen(user string, e *event) {
	// This channel will receive new events to listen to as long as the client
	// reads newer events
	c := make(chan *event)

	next := e
	for {
		// Insert new listener to the list
		next.listeners = append(next.listeners, c)

		// Wait for a new event to listen to
		next = <-c

		log.Println("next", next)

		// Is next event is null, stop listening
		if next == nil {
			// Find the listener index to remove it
			index := -1
			for i, l := range e.listeners {
				if l == c {
					index = i
					break
				}
			}
			if index >= 0 {
				e.listeners = append(e.listeners[:index], e.listeners[index+1:]...)
			}

			// Close channel
			close(c)
			// Terminate goroutine
			return
		}
	}
}

func (b *Events) GetLastEvent(user string) (*backend.Event, error) {
	// No events for this user, create an empty one
	if len(b.events[user]) == 0 {
		err := b.insertEvent(user, &backend.Event{})
		if err != nil {
			return nil, err
		}
	}

	// Get last event
	lastEvent := b.events[user][len(b.events[user])-1]
	// Listen to it
	go b.listen(user, lastEvent)

	return lastEvent.Event, nil
}

func (b *Events) GetEventsAfter(user, id string) (*backend.Event, error) {
	log.Println("events:")
	for i, e := range b.events[user] {
		log.Println(i, e)
	}

	var merged *backend.Event
	var listener chan *event
	from := -1
	last := len(b.events[user]) - 1
	cleanupUntil := -1
	for i, e := range b.events[user] {
		if e.ID == id {
			// This is the event we're looking for

			if len(e.listeners) == 0 {
				// No listener on this event
				// That means that the listener has timed out
				// Create a new one
				return b.GetLastEvent(user)
			}

			// If there were no new events, there's no need to remove the listener
			if i < last {
				listener = e.listeners[0]
				e.listeners = e.listeners[1:]
			}

			from = i
			merged = &backend.Event{ID: e.ID}
		}

		// If there are no listeners on this event anymore, there's no need to keep
		// it in memory, we can destroy it
		// Make sure this isn't the last event, because we're going to add a listener
		// on it
		if cleanupUntil == i-1 && i < last && len(e.listeners) == 0 {
			cleanupUntil = i
		}

		// If this is a new event, merge it
		if from != -1 && i > from {
			merged = mergeEvents(merged, e.Event)
		}
	}

	// Cleanup old events
	if cleanupUntil != -1 {
		b.events[user] = b.events[user][cleanupUntil+1:]
	}

	// Event ID not found
	if from == -1 {
		return nil, errors.New("No such event")
	}

	// If we removed the listener from an event, we have to add it to the last one
	if from < last {
		listener <- b.events[user][len(b.events[user])-1]
	}

	return merged, nil
}

func (b *Events) DeleteAllEvents(user string) error {
	// Stop all listeners
	for _, e := range b.events[user] {
		for _, l := range e.listeners {
			l <- nil
		}
	}

	delete(b.events, user)
	return nil
}

func NewEvents() backend.EventsBackend {
	return &Events{
		events: map[string][]*event{},
	}
}
