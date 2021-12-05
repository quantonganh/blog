package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/pkg/errors"
)

// ComputeHmac256 computes HMAC-SHA256
func ComputeHmac256(message, secret string) (string, error) {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(message))
	if err != nil {
		return "", errors.Wrap(err, "hmac.Write")
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
