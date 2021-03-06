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
	"sync"
	"time"

	simplejson "github.com/bitly/go-simplejson"
)

var _ Cache = &MemoryCache{}

// MemoryItem represents a memory cache item.
type MemoryItem struct {
	val     interface{}
	created int64
	expire  int64
}

func (item *MemoryItem) isExpired() bool {
	return item.expire > 0 &&
		(time.Now().Unix()-item.created) >= item.expire
}

// MemoryCache represents a memory cache adapter implementation.
type MemoryCache struct {
	lock     sync.RWMutex
	items    map[string]*MemoryItem
	interval int // GC interval.
}

// NewMemoryCache creates and returns a new memory cacher.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: make(map[string]*MemoryItem)}
}

// Put puts value into cache with key and expire time.
// If expired is 0, it will be deleted by next GC operation.
func (c *MemoryCache) Put(key string, val interface{}, expire int64) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items[key] = &MemoryItem{
		val:     val,
		created: time.Now().Unix(),
		expire:  expire,
	}
	return nil
}

// Get gets cached value by given key.
func (c *MemoryCache) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil
	}
	if item.isExpired() {
		go c.Delete(key)
		return nil
	}
	return item.val
}

// Delete deletes cached value by given key.
func (c *MemoryCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.items, key)
	return nil
}

// Incr increases cached int-type value by given key as a counter.
func (c *MemoryCache) Incr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}

	item.val, err = Incr(item.val)
	return err
}

// Decr decreases cached int-type value by given key as a counter.
func (c *MemoryCache) Decr(key string) (err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return errors.New("key not exist")
	}

	item.val, err = Decr(item.val)
	return err
}

// IsExist returns true if cached value exists.
func (c *MemoryCache) IsExist(key string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if v, ok := c.items[key]; ok {
		return !v.isExpired()
	}
	return false
}

// Flush deletes all cached data.
func (c *MemoryCache) Flush() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items = make(map[string]*MemoryItem)
	return nil
}

func (c *MemoryCache) checkRawExpiration(key string) {
	item, ok := c.items[key]
	if !ok {
		return
	}

	if item.isExpired() {
		delete(c.items, key)
	}
}

func (c *MemoryCache) startGC() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.interval < 1 {
		return
	}

	if c.items != nil {
		for key := range c.items {
			c.checkRawExpiration(key)
		}
	}

	time.AfterFunc(time.Duration(c.interval)*time.Second, func() { c.startGC() })
}

// StartAndGC starts GC routine based on config string settings.
func (c *MemoryCache) StartAndGC(config string) error {
	js, err := simplejson.NewJson([]byte(config))
	if err != nil {
		return err
	}

	interval := js.Get("Interval").MustInt(60)

	c.lock.Lock()
	c.interval = interval
	c.lock.Unlock()

	go c.startGC()
	return nil
}

func init() {
	Register("memory", NewMemoryCache())
}
