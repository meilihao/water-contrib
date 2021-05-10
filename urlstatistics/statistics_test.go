package urlstatistics

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meilihao/water"
	. "github.com/smartystreets/goconvey/convey"
)

func TestURLStatistics(t *testing.T) {
	i := 200
	Convey("URLStatistics", t, func() {
		r := water.NewRouter()
		r.Use(URLStatistics())
		r.GET("/a", func(c *water.Context) {
			if i%2 == 0 {
				c.WriteHeader(200)
			} else {
				c.WriteHeader(201)
			}

			i++
		})
		e := r.Handler(water.WithNoFoundHandlers(URLStatistics()))

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/a", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldBeIn, []int{http.StatusOK, 201})
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/a", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldBeIn, []int{http.StatusOK, 201})
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "http://localhost:8080/docs/openapi-ui", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("POST", "http://localhost:8080/docs/openapi", nil)
			So(err, ShouldBeNil)
			e.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)
		}

		m := GetURLStatistics()
		So(len(m.urlmap), ShouldEqual, 3)

		buf := bytes.NewBuffer(nil)

		fmt.Println()

		m.GetMap(buf)
		So(buf.Len(), ShouldNotEqual, 0)
		fmt.Println(buf.String())

		buf.Reset()
		m.JSON(buf)
		So(buf.Len(), ShouldNotEqual, 0)
		fmt.Println(buf.String())
	})
}
