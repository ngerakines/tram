package util

import (
	"crypto/sha1"
	"fmt"
)

func Hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
