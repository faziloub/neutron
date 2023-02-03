package api

import (
	"gopkg.in/macaron.v1"

	"github.com/fazilb93/neutron/backend"
)

type EventResp struct {
	Resp
	*backend.Event
}

func (api *Api) GetEvent(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	eventId := ctx.Params("event")

	event, err := api.backend.GetEventsAfter(userId, eventId)
	if err != nil {
		return
	}

	// Retrieve complete user profile if it has been updated
	if event.User != nil {
		event.User, err = api.getCurrentUser(ctx)
		if err != nil {
			return
		}
	}

	// Client crashes if Notices is null
	if event.Notices == nil {
		event.Notices = []string{}
	}

	// Some messages have been updated
	if len(event.Messages) != 0 {
		for _, event := range event.Messages {
			if event.Message != nil {
				api.populateMessage(userId, event.Message)
			}
		}

		event.MessageCounts, err = api.backend.CountMessages(userId)
		if err != nil {
			return
		}

		event.ConversationCounts, err = api.backend.CountConversations(userId)
		if err != nil {
			return
		}

		event.Total, event.Unread = backend.MessagesTotalFromCounts(event.MessageCounts)

		if event.Total.Locations == nil {
			event.Total.Locations = []*backend.LocationTotal{}
		}
		if event.Total.Labels == nil {
			event.Total.Labels = []*backend.LabelTotal{}
		}
		if event.Unread.Locations == nil {
			event.Unread.Locations = []*backend.LocationTotal{}
		}
		if event.Unread.Labels == nil {
			event.Unread.Labels = []*backend.LabelTotal{}
		}
	}

	ctx.JSON(200, &EventResp{
		Resp:  Resp{Ok},
		Event: event,
	})
	return
}
