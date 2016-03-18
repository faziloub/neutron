package api

import (
	"gopkg.in/macaron.v1"
	"github.com/go-macaron/binding"

	"github.com/emersion/neutron/backend"
)

type RespCode int

const (
	Ok RespCode = 1000
	Batch = 1001

	BadRequest = 400
	Unauthorized = 401
	NotFound = 404

	InternalServerError = 500
)

type Req struct {}

type Resp struct {
	Code RespCode
}

type ErrorResp struct {
	Resp
	Error string
	ErrorDescription string
}

func newErrorResp(err error) *ErrorResp {
	return &ErrorResp{
		Resp: Resp{InternalServerError},
		Error: "unknown_error",
		ErrorDescription: err.Error(),
	}
}

type BatchReq struct {
	Req
	IDs []string
}

type BatchResp struct {
	Resp
	Responses []*BatchRespItem
}

func newBatchResp(items []*BatchRespItem) *BatchResp {
	return &BatchResp{
		Resp: Resp{Batch},
		Responses: items,
	}
}

type BatchRespItem struct {
	ID string
	Response interface{}
}

type Api struct {
	backend backend.Backend
	sessions map[string]string
}

func (api *Api) getUid(ctx *macaron.Context) string {
	uid, ok := ctx.Data["uid"]
	if !ok {
		return ""
	}

	return uid.(string)
}

func (api *Api) getSessionToken(ctx *macaron.Context) string {
	sessionToken, ok := ctx.Data["sessionToken"]
	if !ok {
		return ""
	}

	return sessionToken.(string)
}

func (api *Api) getUserId(ctx *macaron.Context) string {
	sessionToken := api.getSessionToken(ctx)
	if sessionToken == "" {
		return ""
	}

	userId, ok := api.sessions[sessionToken]
	if !ok {
		return ""
	}

	return userId
}

func New(m *macaron.Macaron, backend backend.Backend) {
	api := &Api{
		backend: backend,
		sessions: map[string]string{},
	}

	m.Use(func (ctx *macaron.Context) {
		if appVersion, ok := ctx.Req.Header["X-Pm-Appversion"]; ok {
			ctx.Data["appVersion"] = appVersion[0]
		}
		if apiVersion, ok := ctx.Req.Header["X-Pm-Apiversion"]; ok {
			ctx.Data["apiVersion"] = apiVersion[0]
		}
		if sessionToken, ok := ctx.Req.Header["X-Pm-Session"]; ok {
			ctx.Data["sessionToken"] = sessionToken[0]
		}
		if uid, ok := ctx.Req.Header["X-Pm-Uid"]; ok {
			ctx.Data["uid"] = uid[0]
		}
	})

	m.Group("/auth", func() {
		m.Post("/", binding.Json(AuthReq{}), api.Auth)
		m.Delete("/", api.DeleteAuth)
		m.Post("/cookies", binding.Json(AuthCookiesReq{}), api.AuthCookies)
	})

	m.Group("/users", func() {
		m.Get("/", api.GetCurrentUser)
		m.Post("/", binding.Json(CreateUserReq{}), api.CreateUser)
		m.Get("/direct", api.GetDirectUser)
		m.Get("/available/:username", api.GetUsernameAvailable)
		m.Get("/pubkeys/:email", api.GetPublicKey)
	})

	m.Group("/contacts", func() {
		m.Get("/", api.GetContacts)
		m.Post("/", binding.Json(CreateContactsReq{}), api.CreateContacts)
		m.Delete("/", api.DeleteAllContacts)
		m.Put("/:id", binding.Json(UpdateContactReq{}), api.UpdateContact)
		m.Put("/delete", binding.Json(BatchReq{}), api.DeleteContacts)
	})

	m.Group("/labels", func() {
		m.Get("/", api.GetLabels)
	})

	m.Group("/messages", func() {
		m.Get("/", api.ListMessages)
		m.Get("/:id", api.GetMessage)
		m.Get("/count", api.GetMessagesCount)
		m.Put("/:action(read|unread)", binding.Json(BatchReq{}), api.SetMessagesRead)
		m.Put("/:action(star|unstar)", binding.Json(BatchReq{}), api.SetMessagesStar)
		m.Put("/:label(trash|inbox|spam|archive)", binding.Json(BatchReq{}), api.SetMessagesLabel)
		m.Post("/draft", binding.Json(MessageReq{}), api.CreateDraft)
		m.Put("/draft/:id", binding.Json(MessageReq{}), api.UpdateDraft)
		m.Post("/send/:id", binding.Json(SendMessageReq{}), api.SendMessage)
		m.Put("/delete", binding.Json(BatchReq{}), api.DeleteMessages)
	})

	m.Group("/conversations", func() {
		m.Get("/", api.ListConversations)
		m.Get("/:id", api.GetConversation)
		m.Get("/count", api.GetConversationsCount)
		m.Put("/:action(read|unread)", binding.Json(BatchReq{}), api.SetConversationsRead)
		m.Put("/:action(star|unstar)", binding.Json(BatchReq{}), api.SetConversationsStar)
		m.Put("/:label(trash|inbox|spam|archive)", binding.Json(BatchReq{}), api.SetConversationsLabel)
		m.Put("/delete", binding.Json(BatchReq{}), api.DeleteConversations)
	})

	m.Group("/events", func() {
		m.Get("/:event", api.GetEvent)
	})

	m.Group("/settings", func() {
		m.Put("/display", binding.Json(UpdateUserDisplayNameReq{}), api.UpdateUserDisplayName)
	})

	m.Get("/domains/available", api.GetAvailableDomains)

	m.Post("/bugs/crash", binding.Json(CrashReq{}), api.Crash)
}
