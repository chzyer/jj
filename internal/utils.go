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

func GenUuid() string {
	return string(uuid.NewV4()) + string(uuid.NewV4())
}
