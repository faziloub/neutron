package api

import (
	"errors"
	"time"

	"gopkg.in/macaron.v1"

	"github.com/fazilb93/neutron/backend"
)

func getLabelID(name string) (label string) {
	switch name {
	case "draft":
		label = backend.DraftLabel
	case "trash":
		label = backend.TrashLabel
	case "inbox":
		label = backend.InboxLabel
	case "spam":
		label = backend.SpamLabel
	case "archive":
		label = backend.ArchiveLabel
	}
	return
}

func getMessagesFilter(ctx *macaron.Context) *backend.MessagesFilter {
	return &backend.MessagesFilter{
		Limit:       ctx.QueryInt("Limit"),
		Page:        ctx.QueryInt("Page"),
		Label:       ctx.Query("Label"),
		Keyword:     ctx.Query("Keyword"),
		Address:     ctx.Query("Address"),
		Attachments: (ctx.Query("Attachments") == "1"),
		From:        ctx.Query("From"),
		To:          ctx.Query("To"),
		Begin:       ctx.QueryInt64("Begin"),
		End:         ctx.QueryInt64("End"),
		Sort:        ctx.Query("Sort"),
		Desc:        (ctx.Query("Desc") == "1"),
	}
}

func (api *Api) populateMessage(userId string, msg *backend.Message) {
	if msg.ToList == nil {
		msg.ToList = []*backend.Email{}
	}
	if msg.CCList == nil {
		msg.CCList = []*backend.Email{}
	}
	if msg.BCCList == nil {
		msg.BCCList = []*backend.Email{}
	}
	if msg.Attachments == nil {
		msg.Attachments = []*backend.Attachment{}
	}
	if msg.LabelIDs == nil {
		msg.LabelIDs = []string{}
	}

	if msg.Sender != nil {
		msg.SenderAddress = msg.Sender.Address
		msg.SenderName = msg.Sender.Name
	}
	if msg.ReplyTo == nil {
		msg.ReplyTo = msg.Sender
	}

	if backend.IsEncrypted(msg.Body) {
		msg.IsEncrypted = 1
	}

	msg.NumAttachments = len(msg.Attachments)
	if len(msg.Attachments) > 0 {
		msg.HasAttachment = 1
	}

	// TODO: optimize this
	if msg.AddressID == "" && msg.Sender != nil {
		addrs, _ := api.backend.ListAddresses(userId)
		for _, addr := range addrs {
			if addr.Email == msg.Sender.Address {
				msg.AddressID = addr.ID
				break
			}
		}

		if msg.AddressID == "" && len(addrs) > 0 {
			// No address found, pick the first one
			// TODO: pick the main one
			msg.AddressID = addrs[0].ID
		}
	}
}

type MessageReq struct {
	Req
	Message  *backend.Message
	ID       string `json:"id"`
	ParentID string
}

func (req MessageReq) getMessage() *backend.Message {
	return &backend.Message{
		ID:        req.ID,
		ToList:    req.Message.ToList,
		CCList:    req.Message.CCList,
		BCCList:   req.Message.BCCList,
		Subject:   req.Message.Subject,
		IsRead:    req.Message.IsRead,
		AddressID: req.Message.AddressID,
		Body:      req.Message.Body,
	}
}

type MessageResp struct {
	Resp
	Message *backend.Message
}

func (api *Api) GetMessage(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	msgId := ctx.Params("id")

	msg, err := api.backend.GetMessage(userId, msgId)
	if err != nil {
		return
	}

	api.populateMessage(userId, msg)

	ctx.JSON(200, &MessageResp{
		Resp:    Resp{Ok},
		Message: msg,
	})
	return
}

type MessagesListResp struct {
	Resp
	Total    int
	Messages []*backend.Message
}

func (api *Api) ListMessages(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	filter := getMessagesFilter(ctx)

	msgs, total, err := api.backend.ListMessages(userId, filter)
	if err != nil {
		return
	}

	if msgs == nil {
		msgs = []*backend.Message{}
	}

	for _, msg := range msgs {
		api.populateMessage(userId, msg)
	}

	ctx.JSON(200, &MessagesListResp{
		Resp:     Resp{Ok},
		Total:    total,
		Messages: msgs,
	})
	return
}

type MessagesCountResp struct {
	Resp
	Counts []*backend.MessagesCount
}

func (api *Api) GetMessagesCount(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	counts, err := api.backend.CountMessages(userId)
	if err != nil {
		return
	}

	ctx.JSON(200, &MessagesCountResp{
		Resp:   Resp{Ok},
		Counts: counts,
	})
	return
}

type MessagesTotalResp struct {
	Resp
	backend.MessagesTotal
}

func (api *Api) GetMessagesTotal(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	counts, err := api.backend.CountMessages(userId)
	if err != nil {
		return
	}

	totals, _ := backend.MessagesTotalFromCounts(counts)

	if totals.Locations == nil {
		totals.Locations = []*backend.LocationTotal{}
	}
	if totals.Labels == nil {
		totals.Labels = []*backend.LabelTotal{}
	}

	ctx.JSON(200, &MessagesTotalResp{
		Resp:          Resp{Ok},
		MessagesTotal: *totals,
	})
	return
}

type batchMessageUpdater func(*backend.MessageUpdate)

func (api *Api) batchUpdateMessages(ctx *macaron.Context, ids []string, updater batchMessageUpdater) {
	userId := api.getUserId(ctx)

	var respItems []*BatchRespItem

	for _, id := range ids {
		update := &backend.MessageUpdate{
			Message: &backend.Message{ID: id},
		}
		updater(update)

		r := &BatchRespItem{ID: id}
		respItems = append(respItems, r)

		_, err := api.backend.UpdateMessage(userId, update)
		if err != nil {
			r.Response = newErrorResp(err)
		} else {
			r.Response = &Resp{Ok}
		}
	}

	ctx.JSON(200, newBatchResp(respItems))
}

func batchMessageReadUpdater(ctx *macaron.Context) batchMessageUpdater {
	action := ctx.Params("action")

	value := 0
	if action == "read" {
		value = 1
	}

	return func(update *backend.MessageUpdate) {
		update.IsRead = true
		update.Message.IsRead = value
	}
}

func (api *Api) UpdateMessagesRead(ctx *macaron.Context, req BatchReq) {
	api.batchUpdateMessages(ctx, req.IDs, batchMessageReadUpdater(ctx))
}

func batchMessageStarUpdater(ctx *macaron.Context) batchMessageUpdater {
	action := ctx.Params("action")

	value := 0
	labelsAction := backend.AddLabels
	if action == "star" {
		value = 1
		labelsAction = backend.RemoveLabels
	}

	return func(update *backend.MessageUpdate) {
		update.Starred = true
		update.LabelIDs = labelsAction
		update.Message.LabelIDs = []string{backend.StarredLabel}
		update.Message.Starred = value
	}
}

func (api *Api) UpdateMessagesStar(ctx *macaron.Context, req BatchReq) {
	api.batchUpdateMessages(ctx, req.IDs, batchMessageStarUpdater(ctx))
}

func batchMessageSystemLabelUpdater(ctx *macaron.Context) batchMessageUpdater {
	label := getLabelID(ctx.Params("label"))

	return func(update *backend.MessageUpdate) {
		update.LabelIDs = backend.ReplaceLabels
		update.Message.LabelIDs = []string{label}
	}
}

func (api *Api) UpdateMessagesSystemLabel(ctx *macaron.Context, req BatchReq) {
	api.batchUpdateMessages(ctx, req.IDs, batchMessageSystemLabelUpdater(ctx))
}

type UpdateLabelReq struct {
	Req
	Action  int
	LabelID string
}

type UpdateMessagesLabelReq struct {
	UpdateLabelReq
	MessageIDs []string
}

func batchMessageLabelUpdater(ctx *macaron.Context, req UpdateLabelReq) batchMessageUpdater {
	label := req.LabelID

	var action backend.LabelsOperation
	switch req.Action {
	case 0:
		action = backend.RemoveLabels
	case 1:
		action = backend.AddLabels
	}

	return func(update *backend.MessageUpdate) {
		update.LabelIDs = action
		update.Message.LabelIDs = []string{label}
	}
}

func (api *Api) UpdateMessagesLabel(ctx *macaron.Context, req UpdateMessagesLabelReq) {
	api.batchUpdateMessages(ctx, req.MessageIDs, batchMessageLabelUpdater(ctx, req.UpdateLabelReq))
}

func (api *Api) DeleteAllMessages(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	label := getLabelID(ctx.Params("label"))

	// TODO: add a DeleteAllMessages() method to backend

	filter := &backend.MessagesFilter{
		Label: label,
	}

	msgs, _, err := api.backend.ListMessages(userId, filter)
	if err != nil {
		return
	}

	for _, msg := range msgs {
		err = api.backend.DeleteMessage(userId, msg.ID)
		if err != nil {
			return err
		}
	}

	ctx.JSON(200, &Resp{Ok})
	return
}

func (api *Api) CreateDraft(ctx *macaron.Context, req MessageReq) (err error) {
	userId := api.getUserId(ctx)

	user, err := api.backend.GetUser(userId)
	if err != nil {
		return
	}

	msg := req.getMessage()
	msg.Attachments = []*backend.Attachment{}
	msg.LabelIDs = []string{backend.DraftLabel}
	msg.Time = time.Now().Unix()
	msg.Type = backend.DraftType

	if req.ParentID != "" {
		var parent *backend.Message
		parent, err = api.backend.GetMessage(userId, req.ParentID)
		if err != nil {
			return
		}

		msg.ConversationID = parent.ConversationID
	}

	for _, address := range user.Addresses {
		if address.ID == msg.AddressID {
			msg.Sender = address.GetEmail()
			break
		}
	}
	if msg.Sender == nil {
		err = errors.New("Invalid sender address")
		return
	}

	msg, err = api.backend.InsertMessage(userId, msg)
	if err != nil {
		return
	}

	api.populateMessage(userId, msg)

	ctx.JSON(200, &MessageResp{
		Resp:    Resp{Ok},
		Message: msg,
	})
	return
}

func (api *Api) UpdateDraft(ctx *macaron.Context, req MessageReq) (err error) {
	userId := api.getUserId(ctx)
	msgId := ctx.Params("id")

	msg := req.getMessage()
	msg.ID = msgId
	msg.Time = time.Now().Unix()

	msg, err = api.backend.UpdateMessage(userId, &backend.MessageUpdate{
		Message:   msg,
		ToList:    true,
		CCList:    true,
		BCCList:   true,
		Subject:   true,
		IsRead:    true,
		AddressID: true,
		Body:      true,
		Time:      true,
	})
	if err != nil {
		return
	}

	api.populateMessage(userId, msg)

	ctx.JSON(200, &MessageResp{
		Resp:    Resp{Ok},
		Message: msg,
	})
	return
}

type SendMessageReq struct {
	Req
	ID             string `json:"id"`
	Packages       []*backend.MessagePackage
	AttachmentKeys []*backend.AttachmentKey
	ClearBody      string
}

type SendMessageResp struct {
	Resp
	Sent *backend.Message
}

func (api *Api) SendMessage(ctx *macaron.Context, req SendMessageReq) (err error) {
	userId := api.getUserId(ctx)
	msgId := ctx.Params("id")

	msg, err := api.backend.GetMessage(userId, msgId)
	if err != nil {
		return
	}

	outgoing := &backend.OutgoingMessage{Message: msg}

	outgoing.Attachments = make([]*backend.OutgoingAttachment, len(msg.Attachments))
	for i, att := range msg.Attachments {
		att, data, err := api.backend.ReadAttachment(userId, att.ID)
		if err != nil {
			return err
		}

		outgoing.Attachments[i] = &backend.OutgoingAttachment{
			Attachment: att,
			Data:       data,
		}
	}

	// Send each package
	for _, pkg := range req.Packages {
		outgoing.MessagePackage = pkg
		err = api.backend.SendMessage(userId, outgoing)
		if err != nil {
			return
		}
	}

	// If clear body is available, send it to recipients without package
	if req.ClearBody != "" {
		// Decrypt attachments
		if len(req.AttachmentKeys) != len(outgoing.Attachments) {
			err = errors.New("AttachmentKeys count doesn't match Attachments count")
			return
		}
		for i, att := range outgoing.Attachments {
			attKey := req.AttachmentKeys[i]

			data, err := attKey.Decrypt(att.Data)
			if err != nil {
				return err
			}

			att.Data = data
		}

		recipients := []*backend.Email{}
		recipients = append(recipients, msg.ToList...)
		recipients = append(recipients, msg.CCList...)
		recipients = append(recipients, msg.BCCList...)

		// Send clear text message to remaining recipients
		for _, email := range recipients {
			alreadySent := false
			for _, pkg := range req.Packages {
				if pkg.Address == email.Address {
					alreadySent = true
					break
				}
			}
			if alreadySent {
				continue
			}

			outgoing.MessagePackage = &backend.MessagePackage{
				Address: email.Address,
				Body:    req.ClearBody,
			}

			err = api.backend.SendMessage(userId, outgoing)
			if err != nil {
				return
			}
		}
	}

	// Move message to Sent folder
	msg, err = api.backend.UpdateMessage(userId, &backend.MessageUpdate{
		Message: &backend.Message{
			ID:       msgId,
			LabelIDs: []string{backend.SentLabel},
			Type:     backend.SentType,
		},
		Type:     true,
		LabelIDs: backend.ReplaceLabels,
	})
	if err != nil {
		return
	}

	api.populateMessage(userId, msg)

	ctx.JSON(200, &SendMessageResp{
		Resp: Resp{Ok},
		Sent: msg,
	})
	return
}

func (api *Api) DeleteMessages(ctx *macaron.Context, req BatchReq) {
	userId := api.getUserId(ctx)

	var respItems []*BatchRespItem

	for _, id := range req.IDs {
		r := &BatchRespItem{ID: id}
		respItems = append(respItems, r)

		err := api.backend.DeleteMessage(userId, id)
		if err != nil {
			r.Response = newErrorResp(err)
		} else {
			r.Response = &Resp{Ok}
		}
	}

	ctx.JSON(200, newBatchResp(respItems))
}
