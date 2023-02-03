package api

import (
	"encoding/base64"
	"strings"

	"gopkg.in/macaron.v1"

	"github.com/fazilb93/neutron/backend"
)

type UserResp struct {
	Resp
	User *backend.User
}

type CreateUserReq struct {
	Req
	Username   string
	Password   string
	Domain     string
	Email      string
	News       bool
	PrivateKey string
	Token      string
	TokenType  string
}

type DirectUserResp struct {
	Resp
	Direct int
}

type UsernameAvailableResp struct {
	Resp
	Available int
}

func populateUser(user *backend.User) {
	if user.Keys == nil {
		addr := user.GetMainAddress()
		user.Keys = addr.Keys
	}

	for _, addr := range user.Addresses {
		populateAddress(addr)
	}

	user.Role = backend.RolePaidAdmin
	user.Subscribed = 1
	user.Private = 1
}

func (api *Api) getCurrentUser(ctx *macaron.Context) (user *backend.User, err error) {
	userId := api.getUserId(ctx)

	user, err = api.backend.GetUser(userId)
	if err != nil {
		return
	}

	err = api.populateCurrentUser(user)
	return
}

func (api *Api) populateCurrentUser(user *backend.User) (err error) {
	addrs, err := api.backend.ListAddresses(user.ID)
	if err != nil {
		return
	}

	for _, addr := range addrs {
		found := false
		for i, a := range user.Addresses {
			if addr.Email == a.Email {
				user.Addresses[i] = addr
				found = true
				break
			}
		}

		if !found {
			user.Addresses = append(user.Addresses, addr)
		}
	}

	for _, addr := range user.Addresses {
		kp, _ := api.backend.GetKeypair(addr.Email)
		if kp != nil {
			addr.Keys = []*backend.Keypair{kp}
		}
	}

	populateUser(user)
	return
}

func (api *Api) GetCurrentUser(ctx *macaron.Context) {
	user, err := api.getCurrentUser(ctx)
	if err != nil {
		ctx.JSON(200, &ErrorResp{
			Resp:             Resp{NotFound},
			Error:            "invalid_user",
			ErrorDescription: err.Error(),
		})
		return
	}

	ctx.JSON(200, &UserResp{
		Resp: Resp{Ok},
		User: user,
	})
}

func (api *Api) CreateUser(ctx *macaron.Context, req CreateUserReq) (err error) {
	// TODO: check req.Token & req.TokenType

	domain, err := api.backend.GetDomainByName(req.Domain)
	if err != nil {
		return
	}

	email := req.Username + "@" + domain.DomainName

	// Insert user

	user, err := api.backend.InsertUser(&backend.User{
		Name:              req.Username,
		NotificationEmail: req.Email,
	}, req.Password)
	if err != nil {
		return
	}

	// Insert address

	addr := &backend.Address{
		DomainID: domain.ID,
		Email:    email,
		Send:     1,
		Receive:  1,
		Status:   1,
		Type:     1,
	}

	addr, err = api.backend.InsertAddress(user.ID, addr)
	if err != nil {
		return
	}
	user.Addresses = []*backend.Address{addr}

	// Insert keypair

	keypair := backend.NewKeypair("", req.PrivateKey)
	keypair, err = api.backend.InsertKeypair(email, keypair)
	if err != nil {
		return
	}

	user.GetMainAddress().Keys = []*backend.Keypair{keypair}
	populateUser(user)

	ctx.JSON(200, &UserResp{
		Resp: Resp{Ok},
		User: user,
	})
	return
}

func (api *Api) GetDirectUser(ctx *macaron.Context) {
	ctx.JSON(200, &DirectUserResp{
		Resp:   Resp{Ok},
		Direct: 1,
	})
}

func (api *Api) GetUsernameAvailable(ctx *macaron.Context) (err error) {
	username := ctx.Params("username")

	available, err := api.backend.IsUsernameAvailable(username)
	if err != nil {
		return
	}

	value := 0
	if available {
		value = 1
	}

	ctx.JSON(200, &UsernameAvailableResp{
		Resp:      Resp{Ok},
		Available: value,
	})
	return
}

func (api *Api) GetPublicKeys(ctx *macaron.Context) (err error) {
	b, err := base64.URLEncoding.DecodeString(ctx.Params("email"))
	if err != nil {
		return
	}

	emails := strings.Split(string(b), ",")

	resp := map[string]interface{}{}
	for _, email := range emails {
		var key string
		key, err = api.backend.GetPublicKey(email)
		if err != nil {
			return
		}

		resp[email] = key
	}

	resp["Code"] = Ok

	ctx.JSON(200, resp)
	return
}
