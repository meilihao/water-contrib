// Copyright 2015 The Tango Authors
// Copyright 2016 The Water Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package session

import (
	"net/http"
	"time"

	"github.com/meilihao/water"
)

// Tracker provide and set sessionid
type Tracker interface {
	Get(ctx *water.Context) (string, error)
	Set(ctx *water.Context, id string)
	Clear(ctx *water.Context)
}

var _ Tracker = NewCookieTracker("session", 0, false, "/", "")

// CookieTracker provide sessionid from cookie
type CookieTracker struct {
	Name   string
	MaxAge int
	Secure bool
	Path   string
	Domain string
}

func NewCookieTracker(name string, maxAge int, secure bool, path, domain string) *CookieTracker {
	if maxAge < 1 {
		maxAge = 100 * 365 * 24 * 60 * 60
	}
	return &CookieTracker{
		Name:   name,
		MaxAge: maxAge,
		Secure: secure,
		Path:   path,
		Domain: domain,
	}
}

func (tracker *CookieTracker) Get(ctx *water.Context) (string, error) {
	cookie, err := ctx.Req.Cookie(tracker.Name)
	//err only  http.ErrNoCookie
	if err != nil {
		return "", nil
	}
	if cookie.Value == "" {
		return "", nil
	}

	return cookie.Value, nil
}

func (tracker *CookieTracker) Set(ctx *water.Context, id string) {
	cookie, _ := ctx.Req.Cookie(tracker.Name)
	if cookie == nil {
		cookie = &http.Cookie{
			Name:     tracker.Name,
			Value:    id,
			Path:     tracker.Path,
			Domain:   tracker.Domain,
			HttpOnly: true,
			Secure:   tracker.Secure,
			MaxAge:   tracker.MaxAge,
		}

		ctx.Req.AddCookie(cookie)
	} else {
		cookie.Value = id
	}
	http.SetCookie(ctx.ResponseWriter, cookie)
}

func (tracker *CookieTracker) Clear(ctx *water.Context) {
	// 因为一个Cookie应当属于一个path与domain，所以删除时，Cookie的这两个属性也必须设置.
	cookie := http.Cookie{
		Name:     tracker.Name,
		Path:     tracker.Path,
		Domain:   tracker.Domain,
		HttpOnly: true,
		Secure:   tracker.Secure,
		Expires:  time.Date(0, 1, 1, 0, 0, 0, 0, time.Local),
		MaxAge:   0,
	}
	http.SetCookie(ctx.ResponseWriter, &cookie)
}

var _ Tracker = NewHeaderTracker("session")

type HeaderTracker struct {
	Name string
}

func NewHeaderTracker(name string) *HeaderTracker {
	return &HeaderTracker{
		Name: name,
	}
}

func (tracker *HeaderTracker) Get(ctx *water.Context) (string, error) {
	val := ctx.Req.Header.Get(tracker.Name)
	return val, nil
}

func (tracker *HeaderTracker) Set(ctx *water.Context, id string) {
	ctx.ResponseWriter.Header().Set(tracker.Name, id)
}

func (tracker *HeaderTracker) Clear(ctx *water.Context) {
}
