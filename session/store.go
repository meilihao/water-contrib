// Copyright 2015 The Tango Authors
// Copyright 2016 The Water Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package session

type Store interface {
	Get(string) *Container
	Set(string, *Container) error
	Del(string) error
	Flush() error
}
