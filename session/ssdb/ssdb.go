// Copyright 2015 The Tango Authors
// Copyright 2016 The Water Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssdbstore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/bitly/go-simplejson"
	"github.com/meilihao/water-contrib/session"
	"github.com/seefan/gossdb"
)

var _ session.Store = &SsdbStore{}

// SsdbStore represents a redis session store implementation.
type SsdbStore struct {
	prefix string
	pool   *gossdb.Connectors
	maxAge int64
}

// NewSsdbStore creates and returns a redis session store.
func New(config string) (*SsdbStore, error) {
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

	ssdb := &SsdbStore{}
	ssdb.pool = pool

	ssdb.prefix = js.Get("Prefix").MustString("sssdb_")
	ssdb.maxAge = js.Get("MaxAge").MustInt64(0)

	client, err := ssdb.pool.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if !client.Ping() {
		return nil, errors.New("session : ssdb error: wrong config.")
	}

	return ssdb, nil
}

func (c *SsdbStore) serialize(value interface{}) ([]byte, error) {
	err := c.registerGobConcreteType(value)
	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(value).Kind() == reflect.Struct {
		return nil, fmt.Errorf("serialize func only take pointer of a struct")
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)

	err = encoder.Encode(&value)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c *SsdbStore) deserialize(byt []byte) (ptr interface{}, err error) {
	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)

	var p interface{}
	err = decoder.Decode(&p)
	if err != nil {
		return
	}

	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Struct {
		var pp interface{} = &p
		datas := reflect.ValueOf(pp).Elem().InterfaceData()

		sp := reflect.NewAt(v.Type(),
			unsafe.Pointer(datas[1])).Interface()
		ptr = sp
	} else {
		ptr = p
	}
	return
}

func (c *SsdbStore) registerGobConcreteType(value interface{}) error {
	t := reflect.TypeOf(value)

	switch t.Kind() {
	case reflect.Ptr:
		v := reflect.ValueOf(value)
		i := v.Elem().Interface()
		gob.Register(i)
	case reflect.Struct, reflect.Map, reflect.Slice:
		gob.Register(value)
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		// do nothing since already registered known type
	default:
		return fmt.Errorf("unhandled type: %v", t)
	}
	return nil
}

// Get gets value by given key in session.
func (s *SsdbStore) Get(id string) *session.Container {
	c, err := s.pool.NewClient()
	if err != nil {
		return nil
	}
	defer c.Close()

	v, err := c.Get(s.prefix + id)
	if err != nil {
		return nil
	}
	if v.IsEmpty() {
		return nil
	}

	if s.maxAge > 0 {
		_, err = c.Expire(s.prefix+id, s.maxAge)
		if err != nil {
			return nil
		}
	}

	value, err := s.deserialize(v.Bytes())
	if err != nil {
		return nil
	}

	return &session.Container{Data: value}
}

// Set sets value to given key in session.
func (s *SsdbStore) Set(id string, container *session.Container) error {
	if container.Data == nil {
		return nil
	}
	if !container.Changed {
		return nil
	}

	bs, err := s.serialize(container.Data)
	if err != nil {
		return err
	}

	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	if s.maxAge > 0 {
		return c.Set(s.prefix+id, bs, s.maxAge)
	} else {
		return c.Set(s.prefix+id, bs)
	}
}

// Delete delete a key from session.
func (s *SsdbStore) Del(id string) error {
	c, err := s.pool.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	return c.Del(s.prefix + id)
}

func (s *SsdbStore) Flush() error {
	return nil
}
