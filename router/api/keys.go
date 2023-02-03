package api

import (
	"gopkg.in/macaron.v1"

	"github.com/faziloub/neutron/backend"
)

type CreatePrivateKeyReq struct {
	AddressID  string
	PrivateKey string
}

func (api *Api) CreatePrivateKey(ctx *macaron.Context, req CreatePrivateKeyReq) {
	userId := api.getUserId(ctx)

	addr, err := api.backend.GetAddress(userId, req.AddressID)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	kp := backend.NewKeypair("", req.PrivateKey)
	_, err = api.backend.InsertKeypair(addr.Email, kp)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	// Insert new event
	event := backend.NewUserEvent(&backend.User{ID: userId})
	api.backend.InsertEvent(userId, event)

	ctx.JSON(200, &Resp{Ok})
	return
}

type UpdateAllPrivateKeysReq struct {
	Password string
	Keys     []*backend.Keypair
}

func (api *Api) UpdateAllPrivateKeys(ctx *macaron.Context, req UpdateAllPrivateKeysReq) {
	userId := api.getUserId(ctx)

	user, err := api.backend.GetUser(userId)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	// Check password
	user, err = api.backend.Auth(user.Name, req.Password)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	for _, kp := range req.Keys {
		email := kp.ID

		_, err := api.backend.UpdateKeypair(email, kp)
		if err != nil {
			ctx.JSON(500, newErrorResp(err))
			return
		}
	}

	// Insert new event
	event := backend.NewUserEvent(&backend.User{ID: userId})
	api.backend.InsertEvent(userId, event)

	ctx.JSON(200, Resp{Ok})
}
