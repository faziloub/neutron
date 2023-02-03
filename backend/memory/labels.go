package memory

import (
	"errors"

	"github.com/fazilb93/neutron/backend"
	"github.com/fazilb93/neutron/backend/util"
)

type Labels struct {
	labels map[string][]*backend.Label
}

func (b *Labels) getLabelIndex(user, id string) (int, error) {
	for i, lbl := range b.labels[user] {
		if lbl.ID == id {
			return i, nil
		}
	}
	return -1, errors.New("No such label")
}

func (b *Labels) ListLabels(user string) (labels []*backend.Label, err error) {
	labels = b.labels[user]
	return
}

func (b *Labels) InsertLabel(user string, label *backend.Label) (*backend.Label, error) {
	label.ID = util.GenerateId()
	label.Order = len(b.labels[user])
	b.labels[user] = append(b.labels[user], label)
	return label, nil
}

func (b *Labels) UpdateLabel(user string, update *backend.LabelUpdate) (*backend.Label, error) {
	i, err := b.getLabelIndex(user, update.Label.ID)
	if err != nil {
		return nil, err
	}

	label := b.labels[user][i]
	update.Apply(label)
	return label, nil
}

func (b *Labels) DeleteLabel(user, id string) error {
	i, err := b.getLabelIndex(user, id)
	if err != nil {
		return err
	}

	labels := b.labels[user]
	b.labels[user] = append(labels[:i], labels[i+1:]...)

	return nil
}

func NewLabels() backend.LabelsBackend {
	return &Labels{
		labels: map[string][]*backend.Label{},
	}
}
