package events

import (
	"github.com/fazilb93/neutron/backend"
)

type Labels struct {
	backend.LabelsBackend
	events backend.EventsBackend
}

func (b *Labels) InsertLabel(user string, label *backend.Label) (*backend.Label, error) {
	label, err := b.LabelsBackend.InsertLabel(user, label)

	if err == nil {
		event := backend.NewLabelDeltaEvent(label.ID, backend.EventCreate, label)
		b.events.InsertEvent(user, event)
	}

	return label, err
}

func (b *Labels) UpdateLabel(user string, update *backend.LabelUpdate) (*backend.Label, error) {
	label, err := b.LabelsBackend.UpdateLabel(user, update)

	if err == nil {
		event := backend.NewLabelDeltaEvent(label.ID, backend.EventUpdate, label)
		b.events.InsertEvent(user, event)
	}

	return label, err
}

func (b *Labels) DeleteLabel(user, id string) error {
	err := b.LabelsBackend.DeleteLabel(user, id)

	if err == nil {
		event := backend.NewLabelDeltaEvent(id, backend.EventDelete, nil)
		b.events.InsertEvent(user, event)
	}

	return err
}

func NewLabels(bkd backend.LabelsBackend, events backend.EventsBackend) backend.LabelsBackend {
	return &Labels{
		LabelsBackend: bkd,
		events:        events,
	}
}
