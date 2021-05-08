// core from https://github.com/go-macaron/macaron
package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/symwalk"
	"github.com/fsnotify/fsnotify"
	"github.com/meilihao/goutil"
	"github.com/meilihao/logx"
	"github.com/meilihao/water"
)

const (
	_CONTENT_TYPE           = "Content-Type"
	_CONTENT_HTML           = "text/html"
	_CONTENT_XHTML          = "application/xhtml+xml"
	_CONTENT_RenderDuration = "X-Render-Duration"
)

var (
	// Provides a temporary buffer to execute templates into and catch errors.
	bufpool = sync.Pool{
		New: func() interface{} { return new(bytes.Buffer) },
	}

	// from https://github.com/search?l=Go&q=reDefineTag&type=Code
	reDefineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	reTemplateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
	DeepMax                      = 5 //默认模板最多可嵌套5层
)

type RenderOption struct {
	// Directory to load templates. Default is "templates".
	Directory string
	// support theme. Default is to "default".
	Theme string
	// Extensions to parse template files from. Defaults are [".tmpl", ".html"].
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Default is [].
	Funcs []template.FuncMap
	// DelimXXX sets the action delimiters to the specified strings
	DelimLeft  string
	DelimRight string
	// Allows changing of output to XHTML instead of HTML, and without charset. Default is "text/html"
	HTMLContentType string
	// watch tpl, Default is false
	IsWatching bool
}

func (opt *RenderOption) Base() string {
	var base string
	if opt.Theme == "" {
		base = filepath.Join(opt.Directory, "default")
	} else {
		base = filepath.Join(opt.Directory, opt.Theme)
	}

	fi, err := os.Stat(base)
	if err != nil {
		panic(fmt.Errorf(`template set dir "%s" is not found with err: %s`, base, err.Error()))
	}
	if !fi.IsDir() {
		panic(fmt.Errorf(`template set dir "%s" is not found`, base))
	}

	return base
}

// TemplateSet represents a template set of type *template.Template.
type TemplateSet struct {
	lock sync.RWMutex

	_default *templateElem
	sets     map[string]*templateElem
}

type templateElem struct {
	Opt  *RenderOption
	Tpls map[string]*template.Template
}

func (r *TemplateSet) Get(setName string) *templateElem {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var e *templateElem
	if setName == water.DEFAULT_TPL_SET_NAME {
		e = r._default
	} else {
		e = r.sets[setName]
	}

	return e
}

func (r *TemplateSet) renderBytes(setName, tplName string, data interface{}) (*templateElem, *bytes.Buffer, error) {
	e := r.Get(setName)
	if e == nil {
		return nil, nil, fmt.Errorf(`template set "%s" is undefined`, setName)
	}

	buf := bufpool.Get().(*bytes.Buffer)
	buf.Reset()

	tpl := e.Tpls[tplName]
	if tpl == nil {
		return nil, nil, fmt.Errorf(`template "%s:%s" is undefined`, setName, tplName)
	}

	if err := tpl.Execute(buf, data); err != nil {
		return nil, nil, err
	}

	return e, buf, nil
}

func (r *TemplateSet) HTMLSet(ctx *water.Context, code int, setName, tplName string, data interface{}) {
	startTime := time.Now()

	e, buf, err := r.renderBytes(setName, tplName, data)
	if err != nil {
		http.Error(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.Header().Set(_CONTENT_TYPE, e.Opt.HTMLContentType)
	ctx.WriteHeader(code)

	if _, err = buf.WriteTo(ctx); err != nil {
		http.Error(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.Header().Set(_CONTENT_RenderDuration, fmt.Sprintf("%d", time.Since(startTime).Milliseconds()))

	bufpool.Put(buf)
}

func compile(opt *RenderOption) *templateElem {
	e := &templateElem{Tpls: make(map[string]*template.Template), Opt: opt}

	base := opt.Base()
	// Walk 会包含遍历的起点
	err := symwalk.Walk(base, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fp == base { // ignore base
			return nil
		}

		// fp: templates/delims/index.html
		// rp: index.html
		rp, err := filepath.Rel(base, fp)
		if err != nil {
			return err
		}

		if !goutil.InSlice(filepath.Ext(rp), opt.Extensions) {
			return nil
		}

		return addTpl(e, rp)
	})
	if err != nil {
		panic(err)
	}

	return e
}

func (r *TemplateSet) Set(opt *RenderOption) *templateElem {
	e := compile(opt)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.sets[opt.Theme] = e

	if opt.Theme == water.DEFAULT_TPL_SET_NAME {
		r._default = e
	}

	return e
}

func NewRender(opts ...*RenderOption) water.Render {
	if len(opts) == 0 {
		panic("no RenderOption")
	}

	ts := &TemplateSet{
		sets: make(map[string]*templateElem),
	}

	for _, opt := range opts {
		if opt.Directory == "" {
			opt.Directory = "templates"
		}

		if len(opt.Extensions) == 0 {
			opt.Extensions = []string{".tmpl", ".html"}
		}

		if opt.HTMLContentType == "" {
			opt.HTMLContentType = _CONTENT_HTML
		}
		opt.HTMLContentType += "; charset=UTF-8"

		ts.Set(opt)

		if water.Status == water.Dev || opt.IsWatching {
			goWatch(ts, opt)
		}
	}

	return ts
}

func addTpl(e *templateElem, fp string) error {
	opt := e.Opt

	content, err := getFileContent(filepath.Join(opt.Base(), fp))
	if err != nil {
		return err
	}
	if hasDefineTag(content) {
		return nil
	}

	// 存储当前模板依赖的嵌套模板
	needTemplate := make(map[string]struct{})
	err = getTemplateTag(fp, 0, opt, &needTemplate)
	if err != nil {
		return err
	}

	allTpl := make([]string, 0)
	allTpl = append(allTpl, filepath.Join(opt.Base(), fp))

	for k := range needTemplate {
		allTpl = append(allTpl, filepath.Join(opt.Base(), k))
	}

	// [Since the templates created by ParseFiles are named by the base names of the argument files, t should usually have the name of one of the (base) names of the files.](https://golang.google.cn/pkg/html/template/#ParseFiles)
	// 用给定的名称name创建一个template，这个name在后面的ParseFiles里必须存在，不然会保存panic(`template: "example" is an incomplete or empty template`)
	// [执行 ParseFiles 方法时，每个文件都会生成一个模板。只有文件基础名与模板名相同时，该文件的内容才会解析到主模板中](https://studygolang.com/articles/25907)
	tmpl := template.New(filepath.Base(fp))
	if opt.DelimLeft != "" || opt.DelimRight != "" {
		tmpl.Delims(opt.DelimLeft, opt.DelimRight)
	}
	for _, funcs := range opt.Funcs {
		tmpl.Funcs(funcs)
	}
	// Bomb out if parse fails. We don't want any silent server starts.
	template.Must(tmpl.ParseFiles(allTpl...))

	e.Tpls[fp] = tmpl

	return nil
}

// 获取嵌套模板
func getTemplateTag(fp string, deep int, opt *RenderOption, m *map[string]struct{}) error {
	if deep >= DeepMax {
		return errors.New("Temlate too deep.")
	}

	content, err := getFileContent(filepath.Join(opt.Base(), fp))
	if err != nil {
		return errors.New(fmt.Sprintf("Read Template(%s) : %v", fp, err))
	}

	for _, raw := range reTemplateTag.FindAllString(content, -1) {
		parsed := reTemplateTag.FindStringSubmatch(raw)
		tagPath := strings.TrimSpace(parsed[1])
		if !goutil.InSlice(filepath.Ext(tagPath), opt.Extensions) {
			continue
		}

		(*m)[tagPath] = struct{}{}
		if err = getTemplateTag(tagPath, deep+1, opt, m); err != nil {
			return err
		}
	}

	return nil
}

// 判断该模板是否是define模板
// 忽略无需存储的define模板
func hasDefineTag(content string) bool {
	for range reDefineTag.FindAllString(content, -1) {
		return true
	}

	return false
}

// 获取需要监控的文件夹列表
// fsnotify不监控子文件夹,需自行处理
func watchList(opt *RenderOption) map[string]string {
	ls := make(map[string]string)

	base := opt.Base()
	err := symwalk.Walk(base, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			ls[fp] = fp
		}

		return nil
	})
	logx.ErrFatal(err)

	return ls
}

func goWatch(ts *TemplateSet, opt *RenderOption) {
	watcher, err := fsnotify.NewWatcher()
	logx.ErrPanic(err)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Write == fsnotify.Write {
					logx.Info("Reload Templates")

					ts.Set(opt)
				}
			case err := <-watcher.Errors:
				logx.ErrError(err)
			}
		}
	}()

	ls := watchList(opt)
	for k := range ls {
		logx.ErrFatal(watcher.Add(k))
	}
}
