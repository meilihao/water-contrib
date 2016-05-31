package render

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/cxr29/log"
	"github.com/facebookgo/symwalk"
	"github.com/fsnotify/fsnotify"
	"github.com/meilihao/goutil"
	"github.com/meilihao/water"
)

var ErrNoTemplate = errors.New("NoTemplate")

// 存储所有模板
type tplMap map[string]*template.Template

func (t tplMap) HTML(ctx *water.Context, tplName string, data map[string]interface{}) error {
	if water.Status == water.Dev {
		tplLock.RLock()
		defer tplLock.RUnlock()
	}

	v, ok := t[tplName]
	if !ok {
		return ErrNoTemplate
	}

	return v.Execute(ctx, data)
}

// 模板配置文件
type TplConfig struct {
	Dir        string   // 模板根文件夹路径(不以"/"结束,推荐使用相对路径)
	Ext        []string // 支持的模板扩展名,需包含小数点
	FuncMap    template.FuncMap
	DelimLeft  string
	DelimRight string
}

var (
	tplLock       sync.RWMutex
	tplStore      tplMap
	reDefineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	reTemplateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
	DeepMax                      = 5 //默认模板最多可嵌套5层
	needTemplate  map[string]string
)

// New 返回一个模板引擎
func New(tplCfg *TplConfig) (Render, error) {
	tplStore = make(tplMap)

	fi, err := os.Stat(tplCfg.Dir)
	if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, errors.New("TplConfig.Dir must be folder")
	}

	if len(tplCfg.Ext) < 1 {
		return nil, errors.New("Empty TplConfig.Ext")
	}

	if err = parseTpl(tplCfg); err != nil {
		return nil, err
	}

	if len(tplStore) < 1 {
		return nil, errors.New("NoTemplateFiles")
	}

	if water.Status == water.Dev {
		newWatcher(tplCfg)
	}

	return tplStore, nil
}

// 开发模式下,监控模板变化
func newWatcher(tplCfg *TplConfig) {
	watcher, err := fsnotify.NewWatcher()
	log.ErrPanic(err)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Write == fsnotify.Write {
					log.Infoln("Reload Templates")
					log.ErrError(parseTpl(tplCfg))
				}
			case err := <-watcher.Errors:
				log.ErrError(err)
			}
		}
	}()

	ls := watchList(tplCfg)
	if len(ls) < 1 {
		log.Fatalln("No Template Floders To Watch")
	}
	for k := range ls {
		log.ErrFatal(watcher.Add(k))
	}
}

// 解析模板
func parseTpl(tplCfg *TplConfig) error {
	if water.Status == water.Dev {
		tplLock.Lock()
		defer tplLock.Unlock()
	}

	clearTplMap()

	// Walk 会包含遍历的起点
	return symwalk.Walk(tplCfg.Dir, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !goutil.InSlice(filepath.Ext(fPath), tplCfg.Ext) {
			return nil
		}

		addTpl(fPath, tplCfg)

		return nil
	})
}

// 向模板引擎添加模板
func addTpl(fPath string, tplCfg *TplConfig) {
	// 去掉无用的路径前缀
	fPath = strings.TrimPrefix(fPath, filepath.Base(tplCfg.Dir)+"/")

	hasDefine, err := hasDefineTag(fPath, tplCfg)
	if err != nil {
		log.Errorln(err)
		return
	}
	// 忽略无需存储的define模板
	if hasDefine {
		return
	}

	// 存储当前模板依赖的嵌套模板
	needTemplate = make(map[string]string)
	err = getTemplateTag(fPath, 0, tplCfg)
	if err != nil {
		log.Errorln(err)
		return
	}

	allTpl := make([]string, 0)
	allTpl = append(allTpl, filepath.Join(tplCfg.Dir, fPath))

	// only html <==> len(needTemplate)==0
	for k := range needTemplate {
		allTpl = append(allTpl, filepath.Join(tplCfg.Dir, k))
	}

	// 模板名需和模板文件名相同
	t, err := template.New(filepath.Base(fPath)).
		Delims(tplCfg.DelimLeft, tplCfg.DelimRight).
		Funcs(tplCfg.FuncMap).
		ParseFiles(allTpl...)
	if err != nil {
		log.Errorf("Parse Template(%v) error : %s", allTpl, err.Error())
		return
	}

	tplStore[fPath] = t
}

// 获取需要监控的文件夹列表
// fsnotify不监控子文件夹,需自行处理
func watchList(tplCfg *TplConfig) map[string]string {
	ls := make(map[string]string)

	err := symwalk.Walk(tplCfg.Dir, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fInfo.IsDir() {
			ls[fPath] = fPath
		}

		return nil
	})
	log.ErrFatal(err)

	return ls
}

// 清空模板引擎
func clearTplMap() {
	for k := range tplStore {
		delete(tplStore, k)
	}
}

// 判断该模板是否是define模板
// define 以相对于模板根文件夹位置的路劲+模板文件名的形式来命名
func hasDefineTag(fPath string, tplCfg *TplConfig) (bool, error) {
	content, err := getFileContent(filepath.Join(tplCfg.Dir, fPath))
	if err != nil {
		return false, errors.New(fmt.Sprintf("Read Template(%s) : %v", fPath, err))
	}

	for range reDefineTag.FindAllString(content, -1) {
		return true, nil
	}

	return false, nil
}

// 获取嵌套模板
func getTemplateTag(fPath string, deep int, tplCfg *TplConfig) error {
	if deep >= DeepMax {
		return errors.New("Temlate too deep.")
	}

	content, err := getFileContent(filepath.Join(tplCfg.Dir, fPath))
	if err != nil {
		return errors.New(fmt.Sprintf("Read Template(%s) : %v", fPath, err))
	}

	for _, raw := range reTemplateTag.FindAllString(content, -1) {
		parsed := reTemplateTag.FindStringSubmatch(raw)
		tagPath := parsed[1]
		if !goutil.InSlice(filepath.Ext(tagPath), tplCfg.Ext) {
			continue
		}

		needTemplate[tagPath] = tagPath
		err = getTemplateTag(tagPath, deep+1, tplCfg)
		if err != nil {
			return err
		}
	}

	return nil
}
