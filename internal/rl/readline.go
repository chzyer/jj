package rl

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/bobappleyard/readline"
	"github.com/jj-io/jj/internal"
)

var (
	prompt  string
	fname   string
	reading = false
)

func Init() {
	_, f, _, _ := runtime.Caller(1)
	fname = "/tmp/readline." + internal.MD5([]byte(f))
	if b, err := ioutil.ReadFile(fname); err == nil {
		lines := strings.Split(string(b), "\n")
		if len(lines) > 100 {
			lines = lines[len(lines)-100:]
		}
		for _, i := range lines {
			if i != "" {
				readline.AddHistory(i)
			}
		}
	}
}

func Readlinef(fmt_ string, o ...interface{}) string {
	return Readline(fmt.Sprintf(fmt_, o...))
}

func Readline(prompt string) string {
	reading = true
	str, err := readline.String(prompt)
	reading = false
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	readline.AddHistory(str)
	if fname != "" {
		readline.SaveHistory(fname)
	}
	return str
}

func Printf(fmt_ string, o ...interface{}) {
	Println(fmt.Sprintf(fmt_, o...))
}

func Println(str ...interface{}) {
	if reading {
		fmt.Println()
	}
	fmt.Println(str...)
	if reading {
		readline.RefreshLine()
	}
}
