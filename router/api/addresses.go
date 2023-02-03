package api

import (
	"errors"

	"github.com/fazilb93/neutron/backend"
	"gopkg.in/macaron.v1"
)

func populateAddress(addr *backend.Address) {
	if len(addr.Keys) > 0 {
		addr.HasKeys = 1
	}
	if addr.Keys == nil {
		addr.Keys = []*backend.Keypair{}
	}
}

type CreateAddressReq struct {
	Req
	Domain   string
	Local    string
	MemberID string
}

type AddressResp struct {
	Resp
	Address *backend.Address
}

func (api *Api) CreateAddress(ctx *macaron.Context, req CreateAddressReq) (err error) {
	userId := api.getUserId(ctx)

	domain, err := api.backend.GetDomainByName(req.Domain)
	if err != nil {
		return
	}

	email := req.Local + "@" + req.Domain

	addr := &backend.Address{
		DomainID: domain.ID,
		Email:    email,
		Send:     1,
		Receive:  1,
		Status:   1,
		Type:     2,
	}

	addr, err = api.backend.InsertAddress(userId, addr)
	if err != nil {
		return
	}

	populateAddress(addr)

	ctx.JSON(200, &AddressResp{
		Resp:    Resp{Ok},
		Address: addr,
	})
	return
}

func (api *Api) ToggleAddress(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	addrId := ctx.Params("id")
	action := ctx.Params("action")

	update := &backend.AddressUpdate{
		Address: &backend.Address{ID: addrId},
		Status:  true,
	}

	switch action {
	case "enable":
		update.Address.Status = 1
	case "disable":
		update.Address.Status = 0
	default:
		return errors.New("Invalid action")
	}

	_, err = api.backend.UpdateAddress(userId, update)
	if err != nil {
		return
	}

	ctx.JSON(200, &Resp{Ok})
	return
}

func (api *Api) DeleteAddress(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	addrId := ctx.Params("id")

	err = api.backend.DeleteAddress(userId, addrId)
	if err != nil {
		return
	}

	ctx.JSON(200, &Resp{Ok})
	return
}
