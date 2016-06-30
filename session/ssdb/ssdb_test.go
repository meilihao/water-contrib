// Copyright 2015 The Tango Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssdbstore

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meilihao/water"
	"github.com/meilihao/water-contrib/session"
	. "github.com/smartystreets/goconvey/convey"
)

var manager *session.Options

func init() {
	log.SetFlags(log.Lshortfile)

	manager = new(session.Options)
	manager.Generator = session.NewSha1Generator("sha1")
	manager.Tracker = session.NewCookieTracker("session", 0, false, "/", "")
	manager.OnSessionNew = func(s *session.Session) {
		fmt.Println("OnSessionNew")
	}
	manager.OnSessionRelease = func(s *session.Session) {
		fmt.Println("OnSessionRelease")
	}
	store, err := New(`
{
    "SSDB":{
        "Host":"127.0.0.1",
        "Port":8888,
        "MinPoolSize":5,
        "MaxPoolSize":50,
        "AcquireIncrement":5
    },
    "Prefix":"sssdb_",
    "MaxAge":0
}`)
	if err != nil {
		log.Fatalln(err)
	}
	manager.Store = store
}

func Test_Session(t *testing.T) {
	Convey("use session middleware", t, func() {
		router := water.Classic()
		router.Before(session.New(manager))
		router.Get("/", func(ctx *water.Context) {})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		router.ServeHTTP(resp, req)
	})
	Convey("Basic operation", t, func() {
		type Sessdata struct {
			UserName   string
			CreateTime time.Time
		}
		router := water.Classic()
		router.Before(session.New(manager))
		router.Get("/", func(ctx *water.Context) {
			sess := session.Get(ctx)

			sd, ok := sess.Container.Data.(*Sessdata)
			if !ok {
				sd = &Sessdata{}
				sd.UserName = "chen"
				sd.CreateTime = sess.Container.CreateTime

				sess.Container.Data = sd
				sess.Container.Changed = true
			}
		})
		router.Get("/reg", func(ctx *water.Context) {
			sess := session.Get(ctx)

			So(sess.Container.Data, ShouldNotBeNil)
			So(sess.Container.Data.(*Sessdata).UserName, ShouldEqual, "chen")
		})
		router.Get("/get", func(ctx *water.Context) {
			sess := session.Get(ctx)
			sid := sess.Id
			So(sid, ShouldNotBeEmpty)

			sd := sess.Container.Data.(*Sessdata)
			So(sd, ShouldNotBeNil)

			uname := sd.UserName
			So(uname, ShouldEqual, "chen")

			So(sess.Del(sess.Id), ShouldBeNil)

			sc := sess.Get(sess.Id)
			So(sc.Data, ShouldBeNil)
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		So(err, ShouldBeNil)
		router.ServeHTTP(resp, req)

		cookie := resp.Header().Get("Set-Cookie")
		log.Println(cookie)
		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/reg", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		router.ServeHTTP(resp, req)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/get", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		router.ServeHTTP(resp, req)
	})
}
