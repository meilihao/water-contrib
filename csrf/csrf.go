// Copyright 2016 The Water Authors

package csrf

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/cxr29/log"
	"github.com/meilihao/water"
	"github.com/meilihao/water-contrib/session"
	"github.com/seefan/gossdb"
)

var (
	_csrfer csrfer
)

type csrfer interface {
	// generate token to store
	GenerateToken(*water.Context)
	// get token from page
	Token(*water.Context) string
	ValidateToken(*water.Context, string) bool
	Error(*water.Context)
}

func GenerateToken(ctx *water.Context) {
	_csrfer.GenerateToken(ctx)
}

func ValidateToken(ctx *water.Context) {
	var ok bool

	token := _csrfer.Token(ctx)
	if token != "" && _csrfer.ValidateToken(ctx, token) {
		ok = true
	}

	if !ok {
		_csrfer.Error(ctx)
	}
}

type csrf struct {
	// csrf location : Header|Form
	From string
	// csrf name
	Name string
	// to generato token
	TokenGenerator TokenGenerator
	// csrf's store
	Store Store
	// handle invalid token
	ErrorFunc func(*water.Context)
}

type TokenGenerator interface {
	Gen() string
}

type Store interface {
	Get(key string) string
	Set(key, value string) bool
}

func (c *csrf) GenerateToken(ctx *water.Context) {
	token := c.TokenGenerator.Gen()

	if ok := c.Store.Set(session.Get(ctx).Id+"_"+ctx.Req.URL.Path, token); !ok {
		token = ""
	}

	if c.From == "Header" {
		ctx.ResponseWriter.Header().Add(c.Name, token)
	} else {
		ctx.Environ.Set(c.Name, token)
	}
}

func (c *csrf) Token(ctx *water.Context) string {
	switch c.From {
	case "Form":
		return ctx.Req.FormValue(c.Name)
	default:
		return ctx.Req.Header.Get(c.Name)
	}
}

func (c *csrf) ValidateToken(ctx *water.Context, token string) bool {
	tmp := c.Store.Get(session.Get(ctx).Id + "_" + ctx.Req.URL.Path)
	if tmp == "" {
		return false
	}

	return tmp == token
}

func (c *csrf) Error(ctx *water.Context) {
	c.ErrorFunc(ctx)
}

func Init(config string, errorFunc func(*water.Context), store Store, tg TokenGenerator) {
	_csrfer = newCsrf(config, errorFunc, store, tg)
}

func newCsrf(config string, errorFunc func(*water.Context), store Store, tg TokenGenerator) csrfer {
	if errorFunc == nil || store == nil || tg == nil {
		log.Fatalln("csrf : wrong ErrorFunc,Store or TokenGenerator")
	}
	js, err := simplejson.NewJson([]byte(config))
	if err != nil {
		log.Fatalln("csrf : wrong config : ", err)
	}

	c := new(csrf)
	c.ErrorFunc = errorFunc
	c.Store = store
	c.TokenGenerator = tg
	c.From = js.Get("From").MustString("Header")
	c.Name = js.Get("Name").MustString("CrsfToken")

	return c
}

type defaultTokenGenerator struct {
}

func (g *defaultTokenGenerator) Gen() string {
	r := make([]byte, 8)
	io.ReadFull(rand.Reader, r)

	return hex.EncodeToString(r) + fmt.Sprintf("%08x", time.Now().UnixNano())
}

func NewDefaultTokenGenerator() TokenGenerator {
	tg := new(defaultTokenGenerator)

	return tg
}

type defaulStore struct {
	pool *gossdb.Connectors
	// csrf's prefix in store
	prefix string
	// csrf's expire time
	maxAge int64
}

func NewDefaultStore(config string) (Store, error) {
	js, err := simplejson.NewJson([]byte(config))
	if err != nil {
		return nil, err
	}

	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             js.Get("SSDB").Get("Host").MustString(""),
		Port:             js.Get("SSDB").Get("Port").MustInt(0),
		MinPoolSize:      js.Get("SSDB").Get("MinPoolSize").MustInt(0),
		MaxPoolSize:      js.Get("SSDB").Get("MaxPoolSize").MustInt(0),
		AcquireIncrement: js.Get("SSDB").Get("AcquireIncrement").MustInt(0),
	})
	if err != nil {
		return nil, err
	}

	return NewDefaultStoreByInstance(pool, js.Get("Prefix").MustString("csrf_"), js.Get("MaxAge").MustInt64(24*3600))
}

func NewDefaultStoreByInstance(pool *gossdb.Connectors, prefix string, maxAge int64) (Store, error) {
	ssdbStore := &defaulStore{}
	ssdbStore.pool = pool

	client, err := ssdbStore.pool.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if !client.Ping() {
		return nil, errors.New("csrf : ssdb error: wrong config.")
	}

	ssdbStore.prefix = prefix
	ssdbStore.maxAge = maxAge

	if ssdbStore.prefix == "" {
		return nil, errors.New("csrf : empty prefix")
	}

	return ssdbStore, nil
}

func (s *defaulStore) Get(key string) string {
	c, err := s.pool.NewClient()
	if err != nil {
		log.Errorln(err)
		return ""
	}
	defer c.Close()

	v, err := c.Get(s.prefix + key)
	if err != nil {
		log.Errorln(err)
		return ""
	}

	return v.String()
}

func (s *defaulStore) Set(key, value string) bool {
	c, err := s.pool.NewClient()
	if err != nil {
		log.Errorln(err)
		return false
	}
	defer c.Close()

	err = c.Set(s.prefix+key, value, s.maxAge)
	if err != nil {
		log.Errorln(err)
		return false
	}

	return true
}
