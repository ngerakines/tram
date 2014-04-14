package app

import (
	"math/rand"
	"time"
	"crypto/sha1"
	"fmt"
)


type UidManager struct {
	randPool rand.Source
}

func (um *UidManager) Generate() []byte {
	cargo := make([]byte, 16, 16)
	todo := len(cargo)
	offset := 0
	for {
		val := int64(um.randPool.Int63())
		for i := 0; i < 8; i++ {
			cargo[offset] = byte(val & 0xff)
			todo--
			if todo == 0 {
				return cargo
			}
			offset++
			val >>= 8
		}
	}
	panic("unreachable")
}

func (um *UidManager) GenerateHex() string {
	return fmt.Sprintf("%x", um.Generate())
}

func NewUidManager() *UidManager {
	return &UidManager{
		rand.NewSource(int64(time.Now().Nanosecond())),
	}
}

func hash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
