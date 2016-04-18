package render

import (
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cxr29/log"
	"github.com/facebookgo/symwalk"
	"github.com/fsnotify/fsnotify"
	"github.com/meilihao/cutil"
	"github.com/meilihao/water"
)

var ErrNoTemplate = errors.New("NoTemplate")

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

type TplConfig struct {
	Dir        string
	Ext        []string
	FuncMap    template.FuncMap
	DelimLeft  string
	DelimRight string
}

var (
	tplLock       sync.RWMutex
	tplStore      tplMap
	tplCfgGlobal  *TplConfig
	tplPathPrefix string
)

func New(tplCfg *TplConfig) (Render, error) {
	tplCfgGlobal = tplCfg
	tplPathPrefix = filepath.Base(tplCfgGlobal.Dir) + "/"
	tplStore = make(tplMap)

	fi, err := os.Stat(tplCfgGlobal.Dir)
	if err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, errors.New("TplConfig.Dir must be folder")
	}

	if len(tplCfgGlobal.Ext) < 1 {
		return nil, errors.New("Empty TplConfig.Ext")
	}

	if err = parseTpl(tplCfg); err != nil {
		return nil, err
	}

	if len(tplStore) < 1 {
		return nil, errors.New("NoTemplateFiles")
	}

	if water.Status == water.Dev {
		newWatcher()
	}

	return tplStore, nil
}

func newWatcher() {
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
					parseTpl(tplCfgGlobal)
				}
			case err := <-watcher.Errors:
				log.ErrError(err)
			}
		}
	}()

	ls := waterFloderList(tplCfgGlobal)
	if len(ls) < 1 {
		log.Fatalln("No Template Floders To Watch")
	}
	for k := range ls {
		log.ErrFatal(watcher.Add(k))
	}
}

func parseTpl(tplCfg *TplConfig) error {
	if water.Status == water.Dev {
		tplLock.Lock()
		defer tplLock.Unlock()
	}

	clearTplMap()

	return symwalk.Walk(tplCfgGlobal.Dir, func(fPath string, fInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !cutil.InSlice(filepath.Ext(fPath), tplCfgGlobal.Ext) {
			return nil
		}

		addTpl(fPath, tplCfgGlobal)

		return nil
	})
}

func addTpl(fPath string, tplCfg *TplConfig) {
	fPath = strings.TrimPrefix(fPath, tplPathPrefix)

	// 模板名需和模板文件名相同
	t, err := template.New(filepath.Base(fPath)).
		Delims(tplCfgGlobal.DelimLeft, tplCfgGlobal.DelimRight).
		Funcs(tplCfgGlobal.FuncMap).
		ParseFiles(filepath.Join(tplCfgGlobal.Dir, fPath))
	if err != nil {
		log.Errorf("Parse Template(%s) error : %s", fPath, err.Error())
		return
	}

	tplStore[fPath] = t
}

// 获取需要监控的文件夹列表
// fsnotify不监控子文件夹,需自行处理
func waterFloderList(tplCfg *TplConfig) map[string]string {
	ls := make(map[string]string)

	err := symwalk.Walk(tplCfgGlobal.Dir, func(fPath string, fInfo os.FileInfo, err error) error {
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

func clearTplMap() {
	for k := range tplStore {
		delete(tplStore, k)
	}
}
