package qcc

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"strings"
	"unicode/utf16"
)

var signingCodes = [...]byte{'W', 'l', 'k', 'B', 'Q', 'g', 'f', 'i', 'i', 'r', 'v', '6', 'A', 'K', 'N', 'k', '4', 'L', '1', '8'}

// Sign is the dynamically named request header required by QCC.
type Sign struct {
	HeaderName  string
	HeaderValue string
}

// GenerateSign reproduces the QCC browser client's signing algorithm.
// payload must be the exact JSON string sent in the request body.
func GenerateSign(path, payload, tid string) Sign {
	path = strings.ToLower(path)
	payload = strings.ToLower(payload)
	key := signingKey(path)

	return Sign{
		HeaderName:  hmacSHA512(path+payload, key)[8:28],
		HeaderValue: hmacSHA512(path+"pathString"+payload+tid, key),
	}
}

func signingKey(path string) string {
	codeUnits := utf16.Encode([]rune(path + path))
	key := make([]byte, len(codeUnits))
	for index, codeUnit := range codeUnits {
		key[index] = signingCodes[codeUnit%uint16(len(signingCodes))]
	}
	return string(key)
}

func hmacSHA512(payload, key string) string {
	hash := hmac.New(sha512.New, []byte(key))
	hash.Write([]byte(payload))
	return hex.EncodeToString(hash.Sum(nil))
}
