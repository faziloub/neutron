package api

import (
	"gopkg.in/macaron.v1"

	"github.com/faziloub/neutron/backend"
)

type UpdateUserPasswordReq struct {
	Req
	Password    string
	NewPassword string
}

type UpdateUserSettingsReq struct {
	Req
	*backend.User
	Password string
}

func (api *Api) UpdateUserPassword(ctx *macaron.Context, req UpdateUserPasswordReq) {
	userId := api.getUserId(ctx)

	err := api.backend.UpdateUserPassword(userId, req.Password, req.NewPassword)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	ctx.JSON(200, &Resp{Ok})
	return
}

func (api *Api) updateUserSettings(ctx *macaron.Context, update *backend.UserUpdate, updated *backend.User) {
	updated.ID = api.getUserId(ctx)

	update.User = updated

	err := api.backend.UpdateUser(update)
	if err != nil {
		ctx.JSON(500, newErrorResp(err))
		return
	}

	ctx.JSON(200, &Resp{Ok})
	return
}

func (api *Api) UpdateUserDisplayName(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{DisplayName: true}, req.User)
}

func (api *Api) UpdateUserSignature(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{Signature: true}, req.User)
}

func (api *Api) UpdateUserAutoSaveContacts(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{AutoSaveContacts: true}, req.User)
}

func (api *Api) UpdateUserShowImages(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{ShowImages: true}, req.User)
}

func (api *Api) UpdateUserComposerMode(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{ComposerMode: true}, req.User)
}

func (api *Api) UpdateUserViewLayout(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{ViewLayout: true}, req.User)
}

func (api *Api) UpdateUserMessageButtons(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{MessageButtons: true}, req.User)
}

func (api *Api) UpdateUserTheme(ctx *macaron.Context, req UpdateUserSettingsReq) {
	api.updateUserSettings(ctx, &backend.UserUpdate{Theme: true}, req.User)
}
