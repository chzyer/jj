package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jj-io/jj/service"
	"github.com/jj-io/jj/service/auth"
	"github.com/jj-io/jj/service/op"

	"gopkg.in/logex.v1"
)

var (
	srvs = []*service.ServiceType{
		{Name: auth.Name, New: auth.NewAuthService},
		{Name: op.Name, New: op.NewOpService},
	}
)

func usage() {
	print(fmt.Sprintf(`usage: %v [service ...] [service config]

service: 
	%v

service config:
	[service].[config]=xx

example:
	%[1]v auth op -auth.listen=:1111 -op.listen=:2222
`, os.Args[0], strings.Join(serviceNames(), "\n\t")))
	os.Exit(1)
}

func serviceNames() []string {
	s := make([]string, len(srvs))
	for i := range srvs {
		s[i] = srvs[i].Name
	}
	return s
}

func main() {
	if len(os.Args[1:]) == 1 && os.Args[1] == "-h" {
		usage()
	}

	hasServices := false
	optIdx := -1
	for i, srvName := range os.Args[1:] {
		if srvName[0] == '-' {
			optIdx = i + 1
			break
		}
		found := false
		for _, s := range srvs {
			if s.Name == srvName {
				s.Use = true
				hasServices = true
				found = true
				break
			}
		}
		if !found {
			println(fmt.Sprintf("unknown services '%v'", srvName))
			usage()
		}
	}

	// open all services if none service is specified
	if !hasServices {
		for _, s := range srvs {
			s.Use = true
		}
	}

	if optIdx > 0 {
		for _, opt := range os.Args[optIdx:] {
			for _, s := range srvs {
				prefix := "-" + s.Name + "."
				if strings.HasPrefix(opt, prefix) {
					if !s.Use {
						println("panic: service", s.Name, "is not used")
						os.Exit(1)
					}
					s.Args = append(s.Args, "-"+opt[len(prefix):])
					break
				}
			}
		}
	}

	for _, s := range srvs {
		if s.Use {
			s.Ins = s.New(os.Args[0]+" "+s.Name, s.Args)
			if i, ok := s.Ins.(service.ServiceIniter); ok {
				if err := i.Init(); err != nil {
					logex.Error(err)
					os.Exit(1)
				}

			}
		}
	}

	for _, s := range srvs {
		if s.Use {
			logex.Infof("running service %v", s.Name)
			go s.Ins.Run()
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP)
	<-c
}
