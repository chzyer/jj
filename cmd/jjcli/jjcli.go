package main

import (
	"syscall"
	"time"

	"github.com/chzyer/flagx"
	"github.com/jj-io/jj/httprpc"
	"github.com/jj-io/jj/internal"
	"github.com/jj-io/jj/model"
	"github.com/jj-io/readline"
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
	flagx.Parse(&c)
	return &c
}

func GetEmailAndPassword(c *Config) (email, password string) {
	var err error
	email = c.Email
	for email == "" {
		email = readline.String("email: ")
		if email == "" {
			readline.Exit("bye!")
		}
		if !model.RegexpUserEmail.MatchString(email) {
			readline.Errorf("%v is not a valid email", email)
			email = ""
			continue
		}
	}

	print("password: ")
	pswd, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		readline.Exit(err.Error())
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
			readline.Error(err)
			if !noReg {
				isreg := readline.Stringf("want to register as '%v' ?(Y/n): ", email)
				switch isreg {
				case "y", "Y", "":
					resp, err := call.Register(email, pswd)
					if err != nil {
						readline.Exit(err)
					}
					if resp.Result != 200 {
						readline.Errorf(resp.Reason)
						continue
					}
					return email, resp.Uid, resp.Token, resp.MgrAddr
				default:
					noReg = true
				}
			}
			readline.Info("please re-enter login info.")
			continue
		}
		return email, resp.Uid, resp.Token, resp.MgrAddr
	}
	readline.Exit("bye!")
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
	_ = email
	readline.Info("welcome to jj-cli!")
	mgrCli, err := NewMgrCli(mgrAddr)
	if err != nil {
		readline.Exit(err.Error())
	}
	if err := mgrCli.SendInit(uid, token); err != nil {
		readline.Errorf("unexcept error: %v", err)
		return true
	}

	var cmd string
	for err == nil {
		cmd = readline.String("home» ")
		processMgr(cmd)
	}
	return false
}

func main() {
	readline.Init()

	for run() {
	}
}
