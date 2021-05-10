package openapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meilihao/water"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenapiUI(t *testing.T) {
	Convey("openapi-ui", t, func() {
		r := water.NewRouter()
		OpenapiUI(r, &Option{})
		e := r.Handler()

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi-ui", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		//e.Run()
	})
}

func TestOpenapiEditor(t *testing.T) {
	Convey("openapi-editor", t, func() {
		r := water.NewRouter()
		OpenapiEditor(r, &Option{})
		e := r.Handler()

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi-editor", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		//e.Run()
	})
}
