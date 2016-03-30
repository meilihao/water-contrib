// Copyright 2015 The Tango Authors
// Copyright 2016 The Water Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package session

import (
	"log"
	"time"

	"github.com/meilihao/water"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

type Options struct {
	Store            Store
	Generator        IdGenerator
	Tracker          Tracker
	OnSessionNew     func(*Session)
	OnSessionRelease func(*Session)
}

type Container struct {
	Data       interface{}
	CreateTime time.Time
	LastTime   time.Time
	Changed    bool
}

type Session struct {
	Id string
	*Container

	manager *Options
	ctx     *water.Context
}

func New(opt *Options) water.HandlerFunc {
	return func(ctx *water.Context) {
		sess := &Session{manager: opt, ctx: ctx}
		ctx.Environ.Set("Session", sess)

		sess.init()

		ctx.Next()

		var err error
		sess.Container.LastTime = time.Now()
		if err = sess.manager.Store.Set(sess.Id, sess.Container); err != nil {
			log.Println("session : error(1):" + err.Error())
			return
		}
		if sess.manager.OnSessionRelease != nil {
			sess.manager.OnSessionRelease(sess)
		}
	}
}

func (sess *Session) init() {
	var err error
	if sess.Id, err = sess.manager.Tracker.Get(sess.ctx); err != nil {
		log.Println("session : error(2):" + err.Error())
	}

	if sess.Id == "" || !sess.manager.Generator.IsValid(sess.Id) {
		sess.Id = sess.manager.Generator.Gen(sess.ctx.Req)
		sess.manager.Tracker.Set(sess.ctx, sess.Id)

		sess.Container = newContainer()

		if sess.manager.OnSessionNew != nil {
			sess.manager.OnSessionNew(sess)
		}
	} else {
		sess.Container = sess.manager.Store.Get(sess.Id)
		// session is timeout
		if sess.Container == nil {
			sess.Container = newContainer()
		}
	}
}

func newContainer() *Container {
	return &Container{
		Changed:    true,
		CreateTime: time.Now(),
	}
}

func Get(ctx *water.Context) *Session {
	return ctx.Environ.Get("Session").(*Session)
}

func (sess *Session) Get(id string) *Container {
	c := sess.manager.Store.Get(id)
	if c == nil {
		c = newContainer()
	}
	return c
}

func (sess *Session) Set(id string, c *Container) error {
	return sess.manager.Store.Set(id, c)
}

func (sess *Session) Del(id string) error {
	return sess.manager.Store.Del(id)
}

func (sess *Session) Flush() error {
	return sess.manager.Store.Flush()
}
