package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meilihao/water"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Render_HTML(t *testing.T) {
	Convey("Render with nested HTML", t, func() {
		opt := &RenderOption{}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/foobar", func(c *water.Context) {
			c.HTML(200, "posts/nested.html", nil)
		})

		w := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/foobar", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Header().Get(_CONTENT_TYPE), ShouldEqual, _CONTENT_HTML+"; charset=UTF-8")
		So(resp.Body.String(), ShouldEqual, "nested")
	})

	Convey("Render bad HTML", t, func() {
		opt := &RenderOption{
			Theme:      "tmpl_html",
			Extensions: []string{".tmpl", ".html"},
		}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/1", func(c *water.Context) {
			c.HTMLSet(200, "abc", "1.tmpl", nil)
		})

		w := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusInternalServerError)
		So(resp.Body.String(), ShouldEqual, "template set \"abc\" is undefined\n")
	})
}

func Test_Render_XHTML(t *testing.T) {
	Convey("Render XHTML", t, func() {
		opt := &RenderOption{
			Theme:           "tmpl_html",
			Extensions:      []string{".tmpl", ".html"},
			HTMLContentType: _CONTENT_XHTML,
		}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/1", func(c *water.Context) {
			c.HTMLSet(200, opt.Theme, "1.tmpl", nil)
		})

		w := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/1", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Header().Get(_CONTENT_TYPE), ShouldEqual, _CONTENT_XHTML+"; charset=UTF-8")
	})
}

func Test_Render_Extensions(t *testing.T) {
	Convey("Render with extensions", t, func() {
		opt := &RenderOption{
			Theme:      "tmpl_html",
			Extensions: []string{".tmpl", ".html"},
		}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/1", func(c *water.Context) {
			c.HTMLSet(200, opt.Theme, "1.tmpl", nil)
		})
		r.GET("/2", func(c *water.Context) {
			c.HTMLSet(200, opt.Theme, "2.html", nil)
		})

		w := r.Handler()
		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/1", nil)
			So(err, ShouldBeNil)
			w.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Header().Get(_CONTENT_TYPE), ShouldEqual, _CONTENT_HTML+"; charset=UTF-8")
			So(resp.Body.String(), ShouldEqual, "tmpl")
		}
		{
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/2", nil)
			So(err, ShouldBeNil)
			w.ServeHTTP(resp, req)

			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Header().Get(_CONTENT_TYPE), ShouldEqual, _CONTENT_HTML+"; charset=UTF-8")
			So(resp.Body.String(), ShouldEqual, "html")
		}
	})
}

func Test_Render_Funcs(t *testing.T) {
	Convey("Render with functions", t, func() {
		opt := &RenderOption{
			Theme: "custom_funcs",
			Funcs: []template.FuncMap{
				{
					"myCustomFunc": func() string {
						return "My custom function"
					},
				},
			},
		}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/foobar", func(c *water.Context) {
			c.HTMLSet(200, opt.Theme, "index.html", "jeremy")
		})

		w := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/foobar", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		So(resp.Body.String(), ShouldEqual, "My custom function")
	})
}

func Test_Render_Delimiters(t *testing.T) {
	Convey("Render with delimiters", t, func() {
		opt := &RenderOption{
			DelimLeft:  "{[{",
			DelimRight: "}]}",
			Theme:      "delims",
		}
		render := NewRender(opt)
		water.SetRender(render)

		r := water.Default()
		r.GET("/foobar", func(c *water.Context) {
			c.HTMLSet(200, opt.Theme, "index.html", "jeremy")
		})

		w := r.Handler()

		resp := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/foobar", nil)
		So(err, ShouldBeNil)
		w.ServeHTTP(resp, req)

		So(resp.Code, ShouldEqual, http.StatusOK)
		So(resp.Header().Get(_CONTENT_TYPE), ShouldEqual, _CONTENT_HTML+"; charset=UTF-8")
		So(resp.Body.String(), ShouldEqual, "<h1>Hello jeremy</h1>")
	})
}
