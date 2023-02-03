package api

import (
	"gopkg.in/macaron.v1"

	"github.com/faziloub/neutron/backend"
)

type ContactsResp struct {
	Resp
	Contacts []*backend.Contact
}

type CreateContactsReq struct {
	Contacts []*backend.Contact
}

type ContactResp struct {
	Resp
	Contact *backend.Contact
}

type UpdateContactReq struct {
	Req
	ID    string `json:"id"`
	Name  string
	Email string
}

func (api *Api) GetContacts(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	contacts, err := api.backend.ListContacts(userId)
	if err != nil {
		return
	}

	if contacts == nil {
		contacts = []*backend.Contact{}
	}

	ctx.JSON(200, &ContactsResp{
		Resp:     Resp{Ok},
		Contacts: contacts,
	})
	return
}

func (api *Api) CreateContacts(ctx *macaron.Context, req CreateContactsReq) {
	userId := api.getUserId(ctx)

	var respItems []*BatchRespItem

	for _, contact := range req.Contacts {
		r := &BatchRespItem{}
		respItems = append(respItems, r)

		contact, err := api.backend.InsertContact(userId, contact)
		if err != nil {
			r.Response = &ErrorResp{
				Resp:             Resp{InternalServerError},
				ErrorDescription: err.Error(),
			}
		} else {
			r.Response = &ContactResp{
				Resp:    Resp{Ok},
				Contact: contact,
			}
		}
	}

	ctx.JSON(200, &BatchResp{
		Resp:      Resp{Batch},
		Responses: respItems,
	})
}

func (api *Api) UpdateContact(ctx *macaron.Context, req UpdateContactReq) (err error) {
	userId := api.getUserId(ctx)

	contact, err := api.backend.UpdateContact(userId, &backend.ContactUpdate{
		Contact: &backend.Contact{
			ID:    req.ID,
			Name:  req.Name,
			Email: req.Email,
		},
		Name:  true,
		Email: true,
	})
	if err != nil {
		return
	}

	ctx.JSON(200, &ContactResp{
		Resp:    Resp{Ok},
		Contact: contact,
	})
	return
}

func (api *Api) DeleteContacts(ctx *macaron.Context, req BatchReq) {
	userId := api.getUserId(ctx)

	var respItems []*BatchRespItem

	for _, id := range req.IDs {
		r := &BatchRespItem{}
		respItems = append(respItems, r)

		err := api.backend.DeleteContact(userId, id)
		if err != nil {
			r.Response = &ErrorResp{
				Resp:             Resp{InternalServerError},
				ErrorDescription: err.Error(),
			}
		} else {
			r.Response = &Resp{Ok}
		}
	}

	ctx.JSON(200, &BatchResp{
		Resp:      Resp{Batch},
		Responses: respItems,
	})
}

func (api *Api) DeleteAllContacts(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	err = api.backend.DeleteAllContacts(userId)
	if err != nil {
		return
	}

	ctx.JSON(200, &Resp{Ok})
	return
}
