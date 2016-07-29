// Copyright 2016 The Water Authors

package csrf

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/meilihao/water"
	"github.com/meilihao/water-contrib/session"
	"github.com/meilihao/water-contrib/session/ssdb"
	. "github.com/smartystreets/goconvey/convey"
)

var sessionManager *session.Options
var o *csrf

func init() {
	// first,init session
	sessionStore, err := ssdbstore.New(`
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
		log.Fatalln("init session store err: ", err)
	}

	manager := new(session.Options)
	manager.Generator = session.NewSha1Generator("test")
	manager.Tracker = session.NewCookieTracker("session", 0, false, "/", ".test.com")
	manager.OnSessionNew = nil
	manager.OnSessionRelease = nil
	manager.Store = sessionStore
	sessionManager = manager

	o = DefaultCSRF
	o.Secret = "123456789987654321"
	o.Cookie.Name = "_csrf"
	o.Cookie.Path = "/"
}

// recommend
func Test_Header(t *testing.T) {
	Convey("Validate token from Header", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))
		w.Before(New(o))

		w.Get("/test1-1", func(ctx *water.Context) {
		})
		w.Post("/test1-2", func(ctx *water.Context) {
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1-1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		cookie := getCookieCSRF(resp.Header())
		token := resp.Header().Get(o.Name)
		So(token, ShouldNotBeEmpty)
		So(cookie, ShouldContainSubstring, o.Cookie.Name)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 403)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set(o.Name, "123456")
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 403)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set(o.Name, token)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 200)
	})
}

func Test_Form(t *testing.T) {
	o.From = "Form"
	Convey("Validate token from From", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))
		w.Before(New(o))

		token := ""
		w.Get("/test1-1", func(ctx *water.Context) {
			token = ctx.Environ.GetString("CSRF")
			So(token, ShouldNotEqual, "")
		})
		w.Post("/test1-2", func(ctx *water.Context) {
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1-1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		cookie := getCookieCSRF(resp.Header())
		So(cookie, ShouldContainSubstring, o.Cookie.Name)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 403)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", fmt.Sprintf("/test1-2?%s=%s", o.Name, "123456"), nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 403)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", fmt.Sprintf("/test1-2?%s=%s", o.Name, token), nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set(o.Name, token)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 200)
	})
	o.From = "Header"
}

func Test_TokenChange(t *testing.T) {
	Convey("Validate token change", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))
		w.Before(New(o))

		w.Get("/test1-1", func(ctx *water.Context) {
		})
		w.Post("/test1-2", func(ctx *water.Context) {
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1-1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		token := resp.Header().Get(o.Name)
		session := getCookieSession(resp.Header())

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/test1-1", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", session)
		w.ServeHTTP(resp, req)
		cookie := getCookieCSRF(resp.Header())

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		// 此时验证需要session id
		req.Header.Set("Cookie", session)
		req.Header.Set(o.Name, token)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 200)
	})
}

func Test_ValidateTimeout(t *testing.T) {
	// change const -> var
	TIMEOUT = 3 * time.Second
	Convey("Validate token timeout", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))
		w.Before(New(o))

		w.Get("/test1-1", func(ctx *water.Context) {
		})
		w.Post("/test1-2", func(ctx *water.Context) {
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1-1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		session := getCookieSession(resp.Header())
		token := resp.Header().Get(o.Name)
		time.Sleep(5 * time.Second)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test1-2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", "_csrf=123456;")
		req.Header.Set("Cookie", session)
		req.Header.Set(o.Name, token)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, 403)
	})
}

func getCookieCSRF(h http.Header) string {
	s := h["Set-Cookie"]
	for _, v := range s {
		if strings.Contains(v, o.Cookie.Name) {
			return v
		}
	}

	return ""
}

func getCookieSession(h http.Header) string {
	s := h["Set-Cookie"]
	for _, v := range s {
		if strings.Contains(v, "session") {
			return v
		}
	}

	return ""
}
