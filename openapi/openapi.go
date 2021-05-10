package openapi

import (
	"fmt"

	"github.com/meilihao/water"
)

type Option struct {
	// DocsUrl the url for openapi-ui, default is "/docs/openapi-ui"
	DocsUrl string
	// Url the url for openapi, default is "/docs/openapi"
	Url string
	// File openapi doc, default is "openapi.yaml"
	File string
}

func OpenapiUI(r *water.Router, opt *Option) {
	if !r.IsParent() {
		panic("sub router not allowed : OpenapiUI()")
	}

	if opt.DocsUrl == "" {
		opt.DocsUrl = "/docs/openapi-ui"
	}
	if opt.Url == "" {
		opt.Url = "/docs/openapi"
	}
	if opt.File == "" {
		opt.File = "openapi.yaml"
	}

	r.GET(opt.Url, func(c *water.Context) {
		c.File(opt.File)
	})

	r.GET(opt.DocsUrl, func(c *water.Context) {
		c.HTMLRaw(200, fmt.Sprintf(tmpl, opt.Url))
	})
}

var tmpl = `<!DOCTYPE html>
<html>
<head>
<link type="text/css" rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
<link rel="shortcut icon" href="">
<title>openapi</title>
</head>
<body>
<div id="swagger-ui">
</div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
<!-- "SwaggerUIBundle" is now available on the page -->
<script>
const ui = SwaggerUIBundle({
	url: '%s',
	dom_id: '#swagger-ui',
	presets: [
	SwaggerUIBundle.presets.apis,
	SwaggerUIBundle.SwaggerUIStandalonePreset
	],
	layout: "BaseLayout",
	deepLinking: true,
	showExtensions: true,
	showCommonExtensions: true
})
</script>
</body>
</html>`

func OpenapiEditor(r *water.Router, opt *Option) {
	if !r.IsParent() {
		panic("sub router not allowed : OpenapiEditor()")
	}

	if opt.DocsUrl == "" {
		opt.DocsUrl = "/docs/openapi-editor"
	}
	if opt.Url == "" {
		opt.Url = "/docs/openapi"
	}
	if opt.File == "" {
		opt.File = "openapi.yaml"
	}

	r.GET(opt.Url, func(c *water.Context) {
		c.File(opt.File)
	})

	r.GET(opt.DocsUrl, func(c *water.Context) {
		c.HTMLRaw(200, tmplEditor1+fmt.Sprintf(tmplEditor2, opt.Url))
	})
}

// 避免css属性`100%`在fmt.Sprintf时被使用
var tmplEditor1 = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger Editor</title>
  <style>
  * {
    box-sizing: border-box;
  }
  body {
    font-family: Roboto,sans-serif;
    font-size: 9px;
    line-height: 1.42857143;
    color: #444;
    margin: 0px;
  }

  #swagger-editor {
    font-size: 1.3em;
  }

  .container {
    height: 100%;
    max-width: 880px;
    margin-left: auto;
    margin-right: auto;
  }

  #editor-wrapper {
    height: 100%;
    border:1em solid #000;
    border:none;
  }

  .Pane2 {
    overflow-y: scroll;
  }

  </style>
  <link href="https://cdn.jsdelivr.net/npm/swagger-editor-dist@3/swagger-editor.css" rel="stylesheet">
</head>
`

var tmplEditor2 = `<body>
  <div id="swagger-editor"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-editor-dist@3/swagger-editor-bundle.js"> </script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-editor-dist@3/swagger-editor-standalone-preset.js"> </script>
  <script>
  window.onload = function() {
    // Build a system
    const editor = SwaggerEditorBundle({
	  url: '%s',
      dom_id: '#swagger-editor',
      layout: 'StandaloneLayout',
      presets: [
        SwaggerEditorStandalonePreset
      ]
    })
    
    window.editor = editor
  }
  </script>
</body>
</html>`
