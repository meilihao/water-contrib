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

// Package cache is a middleware that provides the cache management of Water.
package cache

import (
	"fmt"

	"github.com/meilihao/water"
)

const _VERSION = "0.1.0"

// Cache is the interface that operates the cache data.
type Cache interface {
	// Put puts value into cache with key and expire time.
	Put(key string, val interface{}, timeout int64) error
	// Get gets cached value by given key.
	Get(key string) interface{}
	// Delete deletes cached value by given key.
	Delete(key string) error
	// Incr increases cached int-type value by given key as a counter.
	Incr(key string) error
	// Decr decreases cached int-type value by given key as a counter.
	Decr(key string) error
	// IsExist returns true if cached value exists.
	IsExist(key string) bool
	// Flush deletes all cached data.
	Flush() error
	// StartAndGC starts GC routine based on config string settings.
	StartAndGC(config string) error
}

var adapters = make(map[string]Cache)

// Register registers a adapter.
func Register(name string, adapter Cache) {
	if adapter == nil {
		panic("cache: cannot register adapter with nil value")
	}
	if _, dup := adapters[name]; dup {
		panic(fmt.Errorf("cache: cannot register adapter '%s' twice", name))
	}
	adapters[name] = adapter
}

// New Create a new cache driver by adapter name and config string.
// config need to be correct JSON as string: {"interval":360}.
// it will start gc automatically.
func New(adapterName, config string) water.HandlerFunc {
	adapter, ok := adapters[adapterName]
	if !ok {
		panic(fmt.Errorf("cache: unknown adapter '%s'(forgot to import?)", adapterName))
	}

	if config == "" || config == "{}" {
		panic(fmt.Errorf("cache: empty config"))
	}

	if err := adapter.StartAndGC(config); err != nil {
		panic(fmt.Errorf("cache: adapter '%s' with wrong config(%s).", adapterName, config))
	}

	return func(ctx *water.Context) {
		ctx.Environ.Set("Cache", adapter)
	}
}

func Get(ctx *water.Context) Cache {
	return ctx.Environ.Get("Cache").(Cache)
}
