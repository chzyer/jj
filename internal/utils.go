package internal

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"

	"github.com/satori/go.uuid"
)

var (
	randNum = rand.New(rand.NewSource(time.Now().Unix()))
	salt    = []byte("9a219ee6-15d6-11e5-af9c-525424be6a56")
)

func Rand() *rand.Rand {
	return randNum
}

func GenUserToken() string {
	return uuid.NewV4().String()[:32]
}

func GenUserPswd(pswd []byte) string {
	hash := md5.New()
	hash.Write(pswd)
	hash.Write(salt)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
