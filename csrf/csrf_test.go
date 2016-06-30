// Copyright 2016 The Water Authors

package csrf

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/meilihao/water"
	"github.com/meilihao/water-contrib/session"
	"github.com/meilihao/water-contrib/session/ssdb"
	. "github.com/smartystreets/goconvey/convey"
)

var sessionManager *session.Options
var errFunc func(*water.Context)
var store Store

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

	csrfStore, err := NewDefaultStore(`{
    "SSDB":{
        "Host":"127.0.0.1",
        "Port":8888,
        "MinPoolSize":5,
        "MaxPoolSize":50,
        "AcquireIncrement":5
    },
    "Prefix":"csrf_",
    "MaxAge":0
}`)
	if err != nil {
		log.Fatalln("init csrf store err: ", err)
	}
	store = csrfStore

	errFunc = func(ctx *water.Context) {
		ctx.BadRequest()
	}
}

// recommend
func Test_GenerateHeader(t *testing.T) {
	Init(NewCsrf(`{
		"From":"Header",
		"Name":"CrsfToken"
		}`, errFunc, store, NewDefaultTokenGenerator()))
	Convey("Generate token to header", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		// generate token
		// method 1:
		w.Get("/test1", func(ctx *water.Context) {
			GenerateToken(ctx)

			So(ctx.ResponseWriter.Header().Get("CrsfToken"), ShouldNotBeEmpty)
		})
		// method 2:
		// recommend
		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Header().Get("CrsfToken"), ShouldNotBeEmpty)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Header().Get("CrsfToken"), ShouldNotBeEmpty)
	})
}

func Test_GenerateForm(t *testing.T) {
	Init(NewCsrf(`{
		"From":"Form",
		"Name":"CrsfToken"
		}`, errFunc, store, NewDefaultTokenGenerator()))
	Convey("Generate token", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		// generate token
		// method 1:
		w.Get("/test1", func(ctx *water.Context) {
			GenerateToken(ctx)

			So(ctx.Environ.GetString("CrsfToken"), ShouldNotBeEmpty)
		})
		// method 2:
		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
			So(ctx.Environ.GetString("CrsfToken"), ShouldNotBeEmpty)
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
	})
}

func Test_ValidateHeader(t *testing.T) {
	Init(NewCsrf(`{
		"From":"Header",
		"Name":"CrsfToken"
		}`, errFunc, store, NewDefaultTokenGenerator()))
	Convey("Validate using right token from Header", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
		})
		w.Post("/test2", ValidateToken)

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		cookie := resp.Header().Get("Set-Cookie")
		csrf := resp.Header().Get("CrsfToken")

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("CrsfToken", csrf)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Validate using incorrect token from Header", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
		})
		w.Post("/test2", ValidateToken)

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusBadRequest)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)

		cookie := resp.Header().Get("Set-Cookie")
		csrf := resp.Header().Get("CrsfToken")
		fmt.Println("right token : ", csrf)
		csrf = "123"

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("CrsfToken", csrf)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func Test_ValidateForm(t *testing.T) {
	Init(NewCsrf(`{
		"From":"Form",
		"Name":"CrsfToken"
		}`, errFunc, store, NewDefaultTokenGenerator()))
	Convey("Validate using right token from Form", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		csrf := ""
		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
			csrf = ctx.Environ.GetString("CrsfToken")
		})
		w.Post("/test2", ValidateToken)

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)

		cookie := resp.Header().Get("Set-Cookie")

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2?CrsfToken="+csrf, nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)
	})

	Convey("Validate using incorrect token from Form", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		csrf := ""
		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
			csrf = ctx.Environ.GetString("CrsfToken")
		})
		w.Post("/test2", ValidateToken)

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusBadRequest)

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)

		cookie := resp.Header().Get("Set-Cookie")
		fmt.Println("right token : ", csrf)
		csrf = "123"

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2?CrsfToken="+csrf, nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusBadRequest)
	})
}

func Test_ValidateTimeout(t *testing.T) {
	Init(NewCsrf(`{
		"From":"Header",
		"Name":"CrsfToken"
		}`, errFunc, store, NewDefaultTokenGenerator()))
	Convey("Validate using right token but timeout", t, func() {
		w := water.NewRouter()
		w.Before(session.New(sessionManager))

		w.Get("/test2", GenerateToken, func(ctx *water.Context) {
		})
		w.Post("/test2", ValidateToken)

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/test2", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		cookie := resp.Header().Get("Set-Cookie")
		csrf := resp.Header().Get("CrsfToken")

		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("CrsfToken", csrf)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusOK)

		fmt.Println("timeout's sessionId : ", cookie)
		time.Sleep(4 * time.Second)
		resp = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/test2", nil)
		So(err, ShouldBeNil)
		req.Header.Set("Cookie", cookie)
		req.Header.Set("CrsfToken", csrf)
		w.ServeHTTP(resp, req)
		So(resp.Code, ShouldEqual, http.StatusBadRequest)
	})
}
