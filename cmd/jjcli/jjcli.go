package main

import (
	"github.com/jj-io/jj/httprpc"
	"github.com/jj-io/jj/internal"
	"github.com/jj-io/jj/model"
)

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bobappleyard/readline"
	"github.com/chzyer/reflag"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/logex.v1"
)

type Config struct {
	Auth    string        `flag:"def=http://localhost:8681"`
	Email   string        `flag:"email"`
	Timeout time.Duration `flag:"def=10s"`
}

func NewConfig() *Config {
	var c Config
	reflag.Parse(&c)
	return &c
}

func Exit(s string) {
	println(s)
	os.Exit(1)
}

func GetEmailAndPassword(c *Config) (email, password string) {
	var err error
	email = c.Email
	for email == "" {
		email, err = readline.String("email: ")
		if err != nil {
			Exit(err.Error())
		}
		if email == "" {
			Exit("bye!")
		}
		if !model.RegexpUserEmail.MatchString(email) {
			println(fmt.Sprintf("%v is not a valid email", email))
			email = ""
			continue
		}
	}

	print("password: ")
	pswd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		Exit(err.Error())
	}
	println()

	return email, internal.GenUserPswd(pswd)
}

func loginAndGetInfo(call *Call, conf *Config) (email, uid, token string, mgrAddr []string) {
	noReg := false
	for {
		email, pswd := GetEmailAndPassword(conf)
		resp, err := call.Login(email, pswd)
		if err != nil {
			println(err.Error())
			if !noReg {
				isreg, _ := readline.String(fmt.Sprintf("want to register as '%v' ?(Y/n): ", email))
				switch isreg {
				case "y", "Y", "":
					resp, err := call.Register(email, pswd)
					if err != nil {
						Exit(err.Error())
					}
					if resp.Result != 200 {
						println(resp.Reason)
						continue
					}
					return email, resp.Uid, resp.Token, resp.MgrAddr
				default:
					noReg = true
				}
			}
			println("please re-enter login info.")
			continue
		}
		return email, resp.Uid, resp.Token, resp.MgrAddr
	}
	Exit("bye!")
	return
}

func run() bool {
	conf := NewConfig()
	client, err := httprpc.NewClient(conf.Auth, conf.Timeout)
	if err != nil {
		logex.Fatal(err)
	}
	call := NewCall(client)
	email, uid, token, mgrAddr := loginAndGetInfo(call, conf)
	fmt.Println("welcome to jj-cli!")
	mgrCli, err := NewMgrCli(mgrAddr)
	if err != nil {
		Exit(err.Error())
	}
	if err := mgrCli.SendInit(uid, token); err != nil {
		println("unexcept error:", err.Error())
		return true
	}

	var cmd string
	for err == nil {
		cmd, err = readline.String(email + "> ")
		if err != nil {
			continue
		}
		processMgr(cmd)
	}
	return false
}

func main() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
		<-c
		Exit("\nbye")
	}()
	for run() {
	}
}
