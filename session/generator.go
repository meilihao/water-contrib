// Copyright 2015 The Tango Authors
// Copyright 2016 The Water Authors
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package session

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type IdGenerator interface {
	Gen(req *http.Request) string
	IsValid(id string) bool
}

type Sha1Generator struct {
	hashKey string
}

func NewSha1Generator(hashKey string) *Sha1Generator {
	return &Sha1Generator{hashKey: hashKey}
}

var _ IdGenerator = NewSha1Generator("test")

func GenRandKey(strength int) []byte {
	k := make([]byte, strength)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

func (gen *Sha1Generator) Gen(req *http.Request) string {
	bs := GenRandKey(24)
	if len(bs) == 0 {
		return ""
	}

	sig := fmt.Sprintf("%s%d%02x", req.RemoteAddr, time.Now().UnixNano(), bs)

	data := sha1.Sum([]byte(gen.hashKey + sig))
	return strings.TrimSuffix(base64.URLEncoding.EncodeToString(data[:]), "=")
}

func (gen *Sha1Generator) IsValid(id string) bool {
	return len(id) == 27
}
