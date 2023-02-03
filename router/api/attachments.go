package api

import (
	"encoding/base64"
	"io/ioutil"
	"mime/multipart"

	"github.com/faziloub/neutron/backend"
	"gopkg.in/macaron.v1"
)

func (api *Api) GetAttachment(ctx *macaron.Context) (b []byte, err error) {
	userId := api.getUserId(ctx)
	id := ctx.Params("id")

	att, b, err := api.backend.ReadAttachment(userId, id)
	if err != nil {
		return
	}

	if att.KeyPackets != "" {
		ctx.Resp.Header().Set("Content-Type", "application/pgp")
	} else {
		ctx.Resp.Header().Set("Content-Type", att.MIMEType)
	}

	ctx.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+att.Name+"\"")
	ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Resp.Header().Set("Expires", "0")
	ctx.Resp.Header().Set("Cache-Control", "must-revalidate")
	ctx.Resp.Header().Set("Pragma", "public")
	return
}

// form: attributes are needed to parse multipart form
// See https://github.com/go-macaron/binding/issues/10
type UploadAttachmentReq struct {
	Filename   string                `form:"Filename"`
	MessageID  string                `form:"MessageID"`
	MIMEType   string                `form:"MIMEType"`
	ContentID  string                `form:"ContentID"`
	KeyPackets *multipart.FileHeader `form:"KeyPackets"`
	DataPacket *multipart.FileHeader `form:"DataPacket"`
}

type UploadAttachmentResp struct {
	Resp
	AttachmentID string
	Size         int
}

func (api *Api) UploadAttachment(ctx *macaron.Context, req UploadAttachmentReq) (err error) {
	userId := api.getUserId(ctx)

	kpf, err := req.KeyPackets.Open()
	if err != nil {
		return
	}
	defer kpf.Close()

	kp, err := ioutil.ReadAll(kpf)
	if err != nil {
		return
	}

	df, err := req.DataPacket.Open()
	if err != nil {
		return
	}
	defer df.Close()

	data, err := ioutil.ReadAll(df)
	if err != nil {
		return
	}

	att := &backend.Attachment{
		Name:       req.Filename,
		MessageID:  req.MessageID,
		MIMEType:   req.MIMEType,
		KeyPackets: base64.StdEncoding.EncodeToString(kp),
	}

	att, err = api.backend.InsertAttachment(userId, att, data)
	if err != nil {
		return
	}

	ctx.JSON(200, &UploadAttachmentResp{
		Resp:         Resp{Ok},
		AttachmentID: att.ID,
		Size:         att.Size,
	})
	return
}

func (api *Api) DeleteAttachment(ctx *macaron.Context) error {
	userId := api.getUserId(ctx)
	id := ctx.Params("id")

	if err := api.backend.DeleteAttachment(userId, id); err != nil {
		return err
	}

	ctx.JSON(200, &Resp{Ok})
	return nil
}
