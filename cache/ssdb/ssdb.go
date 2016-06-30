// Copyright 2013 Beego Authors
// Copyright 2014 The Macaron Authors
// Copyright 2016 The Water Authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package cache

import (
	"errors"
	"fmt"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/meilihao/water-contrib/cache"
	"github.com/seefan/gossdb"
)

var _ cache.Cache = &SsdbCache{}

// SsdbCache represents a ssdb cache adapter implementation.
type SsdbCache struct {
	pool   *gossdb.Connectors
	prefix string
}

// Put puts value into cache with key and expire time.
// If expired is 0, it lives forever.
func (c *SsdbCache) Put(key string, val interface{}, expire int64) error {
	client, err := c.pool.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if expire > 0 {
		if err = client.Set(c.prefix+key, val, expire); err != nil {
			return err
		}
	} else {
		if err = client.Set(c.prefix+key, val); err != nil {
			return err
		}
	}

	return nil
}

// Get gets cached value by given key.
func (c *SsdbCache) Get(key string) interface{} {
	client, err := c.pool.NewClient()
	if err != nil {
		fmt.Println("cache : ssdb error:" + err.Error())
		return nil
	}
	defer client.Close()

	val, err := client.Get(c.prefix + key)
	if err != nil {
		fmt.Println("cache : ssdb error:" + err.Error())
		return nil
	}

	if val.IsEmpty() { //not_found
		return nil
	}

	return val
}

// Delete deletes cached value by given key.
func (c *SsdbCache) Delete(key string) error {
	client, err := c.pool.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Del(c.prefix + key); err != nil {
		return err
	}

	return nil
}

// Incr increases cached int-type value by given key as a counter.
func (c *SsdbCache) Incr(key string) error {
	client, err := c.pool.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Incr(c.prefix+key, 1)
	return err
}

// Decr decreases cached int-type value by given key as a counter.
func (c *SsdbCache) Decr(key string) error {
	client, err := c.pool.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Incr(c.prefix+key, -1)
	return err
}

// IsExist returns true if cached value exists.
func (c *SsdbCache) IsExist(key string) bool {
	client, err := c.pool.NewClient()
	if err != nil {
		fmt.Println("cache : ssdb error:" + err.Error())
		return false
	}
	defer client.Close()

	if re, err := client.Exists(c.prefix + key); re && err == nil {
		return true
	} else if err != nil {
		fmt.Println("cache : ssdb error:" + err.Error())
	}

	return false
}

// Flush deletes all cached data.
func (c *SsdbCache) Flush() error {
	// manual operation
	return nil
}

// StartAndGC starts GC routine based on config string settings.
// AdapterConfig: {"Host":"xxx",...,"Prefix":"cssdb_"}
func (c *SsdbCache) StartAndGC(config string) error {
	js, err := simplejson.NewJson([]byte(config))
	if err != nil {
		return err
	}

	pool, err := gossdb.NewPool(&gossdb.Config{
		Host:             js.Get("SSDB").Get("Host").MustString(""),
		Port:             js.Get("SSDB").Get("Port").MustInt(0),
		MinPoolSize:      js.Get("SSDB").Get("MinPoolSize").MustInt(0),
		MaxPoolSize:      js.Get("SSDB").Get("MaxPoolSize").MustInt(0),
		AcquireIncrement: js.Get("SSDB").Get("AcquireIncrement").MustInt(0),
	})
	if err != nil {
		return err
	}
	c.pool = pool

	c.prefix = js.Get("Prefix").MustString("cssdb_")

	client, err := c.pool.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if !client.Ping() {
		return errors.New("cache : ssdb error: wrong config.")
	}

	return nil
}

func init() {
	cache.Register("ssdb", &SsdbCache{})
}
