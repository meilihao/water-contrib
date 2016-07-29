// Copyright 2012 Google Inc. All Rights Reserved.
// Copyright 2014 The Macaron Authors
// Copyright 2016 The Water Authors
package csrf

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

const (
	TIMEOUT = 24 * time.Hour
)

func generateToken(key, id string, now time.Time) string {
	h := hmac.New(sha1.New, []byte(key))
	fmt.Fprintf(h, "%s:%d", id, now.UnixNano())
	token := fmt.Sprintf("%s:%d", h.Sum(nil), now.UnixNano())
	return base64.URLEncoding.EncodeToString([]byte(token))
}

func validToken(token, secret, id string) bool {
	// Decode the token.
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	// Extract the issue time of the token.
	sep := bytes.LastIndex(data, []byte{':'})
	if sep < 0 {
		return false
	}
	nanos, err := strconv.ParseInt(string(data[sep+1:]), 10, 64)
	if err != nil {
		return false
	}
	issueTime := time.Unix(0, nanos)

	now := time.Now()
	// Check that the token is not expired.
	if now.Sub(issueTime) >= TIMEOUT {
		return false
	}

	// Check that the token is not from the future.
	// Allow 1 minute grace period in case the token is being verified on a
	// machine whose clock is behind the machine that issued the token.
	if issueTime.After(now.Add(1 * time.Minute)) {
		return false
	}

	expected := generateToken(secret, id, issueTime)

	// Check that the token matches the expected value.
	// Use constant time comparison to avoid timing attacks.
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
