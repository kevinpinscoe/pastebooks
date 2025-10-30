package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
)

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// Compact signed token helpers (NOT full JWT, but OK for this scaffold)
func signCompact(claims any, secret []byte) string {
	b, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(b)
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return payload + "." + sig
}

func parseCompact(tok string, secret []byte) (*jwtClaims, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 2 {
		return nil, Err("bad token")
	}
	payload, sig := parts[0], parts[1]
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	if base64.RawURLEncoding.EncodeToString(h.Sum(nil)) != sig {
		return nil, Err("sig")
	}
	b, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}
	var cl jwtClaims
	if err := json.Unmarshal(b, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
func Err(s string) error          { return simpleErr(s) }
