package api

import (
	"errors"

	"gopkg.in/macaron.v1"

	"github.com/faziloub/neutron/backend"
)

type LabelsResp struct {
	Resp
	Labels []*backend.Label
}

func (api *Api) GetLabels(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)

	labels, err := api.backend.ListLabels(userId)
	if err != nil {
		return
	}

	if labels == nil {
		labels = []*backend.Label{}
	}

	ctx.JSON(200, &LabelsResp{
		Resp:   Resp{Ok},
		Labels: labels,
	})
	return
}

type LabelReq struct {
	*backend.Label
}

type LabelResp struct {
	Resp
	Label *backend.Label
}

func (api *Api) CreateLabel(ctx *macaron.Context, req LabelReq) (err error) {
	userId := api.getUserId(ctx)

	label, err := api.backend.InsertLabel(userId, &backend.Label{
		Name:    req.Name,
		Display: req.Display,
		Color:   req.Color,
	})
	if err != nil {
		return
	}

	ctx.JSON(200, &LabelResp{
		Resp:  Resp{Ok},
		Label: label,
	})
	return
}

func (api *Api) UpdateLabel(ctx *macaron.Context, req LabelReq) (err error) {
	userId := api.getUserId(ctx)

	req.Label.ID = ctx.Params("id")

	label, err := api.backend.UpdateLabel(userId, &backend.LabelUpdate{
		Label:   req.Label,
		Name:    true,
		Display: true,
		Color:   true,
	})
	if err != nil {
		return
	}

	ctx.JSON(200, &LabelResp{
		Resp:  Resp{Ok},
		Label: label,
	})
	return
}

type LabelsOrderReq struct {
	Order []int
}

func (api *Api) UpdateLabelsOrder(ctx *macaron.Context, req LabelsOrderReq) (err error) {
	userId := api.getUserId(ctx)

	labels, err := api.backend.ListLabels(userId)
	if err != nil {
		return
	}

	if len(labels) != len(req.Order) {
		err = errors.New("Bad order length")
		return
	}

	for i, lbl := range labels {
		_, err = api.backend.UpdateLabel(userId, &backend.LabelUpdate{
			Label: &backend.Label{
				ID:    lbl.ID,
				Order: req.Order[i],
			},
			Order: true,
		})
		if err != nil {
			return
		}
	}

	ctx.JSON(200, &Resp{Ok})
	return
}

func (api *Api) DeleteLabel(ctx *macaron.Context) (err error) {
	userId := api.getUserId(ctx)
	labelId := ctx.Params("id")

	err = api.backend.DeleteLabel(userId, labelId)
	if err != nil {
		return err
	}

	ctx.JSON(200, &Resp{Ok})
	return
}
