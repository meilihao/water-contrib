// Copyright 2016 The Water Authors

package csrf

import (
	"net/http"
	"time"

	"github.com/meilihao/goutil"
	"github.com/meilihao/water"
	"github.com/meilihao/water-contrib/session"
)

var (
	ignoreMethods = []string{
		"GET", "HEAD", "OPTIONS", "TRACE",
	}

	DefaultCSRF = &csrf{
		From:   "Header",
		Name:   "X-CSRF",
		Cookie: http.Cookie{},
		NoFunc: func(ctx *water.Context) {
			ctx.Forbidden()
			ctx.WriteString("csrf : no token")
		},
		ErrorFunc: func(ctx *water.Context) {
			ctx.Forbidden()
			ctx.WriteString("csrf : bad token")
		},
	}
)

type csrf struct {
	// csrf location : Header|Form
	From string
	// csrf name
	Name              string
	Cookie            http.Cookie
	Secret            string
	ErrorFunc, NoFunc func(*water.Context)
}

func (c *csrf) Validate(ctx *water.Context) {
	token := ""

	switch c.From {
	case "Form":
		token = ctx.Req.FormValue(c.Name)
	default:
		token = ctx.Req.Header.Get(c.Name)
	}

	if token == "" {
		c.ErrorFunc(ctx)
		return
	}

	if ctx.Cookie(c.Cookie.Name) == token {
		return
	}

	// delete old token
	ctx.SetCookie(c.Cookie.Name, "", -1, c.Cookie.Path, c.Cookie.Domain)
	if !c.ValidateToken(token, session.Get(ctx).Id) {
		c.ErrorFunc(ctx)
	}
}

func (c *csrf) ValidateToken(token, id string) bool {
	return validToken(token, c.Secret, id)
}

func (c *csrf) Generate(ctx *water.Context) {
	token := ctx.Cookie(c.Cookie.Name)
	if token == "" {
		token = c.GenerateToken(session.Get(ctx).Id)
		ctx.SetCookie(c.Cookie.Name, token, int(TIMEOUT/time.Second), c.Cookie.Path, c.Cookie.Domain, false, true)
	}

	if c.From == "Header" {
		ctx.ResponseWriter.Header().Add(c.Name, token)
	} else {
		ctx.Environ.Set("CSRF", token)
	}
}

func (c *csrf) GenerateToken(id string) string {
	return generateToken(c.Secret, id, time.Now())
}

func New(o *csrf) water.HandlerFunc {
	if o == nil {
		panic("csrf2 : csrf need option")
	}
	if len(o.Secret) < 16 {
		panic("csrf2 : csrf need len(secret) >= 16")
	}
	if o.Cookie.Name == "" || o.Cookie.Path == "" {
		panic("csrf2 : csrf need templateCookie")
	}

	return func(ctx *water.Context) {
		if goutil.InSlice(ctx.Req.Method, ignoreMethods) {
			o.Generate(ctx)
		} else {
			o.Validate(ctx)
		}
	}
}
