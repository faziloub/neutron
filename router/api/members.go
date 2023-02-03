package api

import (
	"github.com/fazilb93/neutron/backend"
	"gopkg.in/macaron.v1"
)

type Member struct {
	ID        string
	NickName  string
	Role      int
	Addresses []*backend.Address
	Private   int
}

type MembersResp struct {
	Resp
	Members []*Member
}

func (api *Api) GetMembers(ctx *macaron.Context) {
	userId := api.getUserId(ctx)

	user, err := api.backend.GetUser(userId)
	if err != nil {
		ctx.JSON(200, newErrorResp(err))
		return
	}

	member := &Member{
		ID:        user.ID,
		NickName:  user.Name,
		Role:      backend.RolePaidAdmin,
		Addresses: user.Addresses,
		Private:   1,
	}

	ctx.JSON(200, &MembersResp{
		Resp:    Resp{Ok},
		Members: []*Member{member},
	})
	return
}
