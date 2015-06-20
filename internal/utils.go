package internal

import (
	"math/rand"
	"time"

	"github.com/satori/go.uuid"
)

var (
	randNum = rand.New(rand.NewSource(time.Now().Unix()))
)

func Rand() *rand.Rand {
	return randNum
}

func GenUserToken() string {
	return uuid.NewV4().String()[:32]
}
