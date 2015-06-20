package internal

import (
	"crypto/md5"
	"encoding/base64"
	"math/rand"
	"time"
)

var (
	salt    = []byte("e448579a-142c-11e5-93fa-525424be6a56")
	randNum = rand.New(rand.NewSource(time.Now().Unix()))
)

func Rand() *rand.Rand {
	return randNum
}

func GenUuid(secret []byte) string {
	h := md5.New()
	h.Write(salt)
	h.Write([]byte(secret))
	randByte := make([]byte, 8)
	for i := 0; i < 8; i++ {
		randByte[i] = byte(rand.Intn(256))
	}
	h.Write(randByte)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
