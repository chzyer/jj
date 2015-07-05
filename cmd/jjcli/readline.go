package main

import (
	"fmt"
	"os"

	"github.com/jj-io/jj/internal/rl"
)

func color(col, s string) string {
	if col == "" {
		return s
	}
	return "\x1b[0;" + col + "m" + s + "\x1b[0m"
}

func Exit(s interface{}) {
	rl.Println(s)
	os.Exit(1)
}

func Info(obj ...interface{}) {
	rl.Println(obj...)
}

func Error(err error) {
	rl.Println(color("31", err.Error()))
}

func Errorf(str string, obj ...interface{}) {
	rl.Println(color("31", fmt.Sprintf(str, obj...)))
}
