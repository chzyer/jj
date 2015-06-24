package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/jj-io/jj/model"

	"gopkg.in/logex.v1"
)

var (
	ErrInvalidToken = logex.Define("invalid token")
)

var (
	RouterRegister = "/auth/register"
	RouterLogin    = "/auth/login"
	RouterInit     = "/auth/init"
)

func InitHandler(mux *http.ServeMux) {
	mux.HandleFunc(RouterRegister, Register)
	mux.HandleFunc(RouterLogin, Login)
	mux.HandleFunc(RouterInit, Init)
}

func response(w http.ResponseWriter, obj interface{}) {
	switch result := obj.(type) {
	case error:
		logex.Error(result)
		w.Header().Set("Content-Type", "application/json")
		data, err := json.Marshal(&ErrorResp{
			Result: 400,
			Reason: result.Error(),
		})
		if err != nil {
			logex.Error(err)
			break
		}
		w.Write(data)
	case url.Values:
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		io.WriteString(w, result.Encode())
	default:
		w.Header().Set("Content-Type", "application/json")
		data, err := json.Marshal(obj)
		if err != nil {
			logex.Error(err)
			break
		}
		w.Write(data)
	}
}

type RegisterResp struct {
	Result     int      `json:"result"`
	Reason     string   `json:"reason,omitempty"`
	Token      string   `json:"token"`
	Uid        string   `json:"uid"`
	MgrAddr    []string `json:"mgraddr"`
	NotifyAddr []string `json:"notifyaddr"`
}

type ErrorResp struct {
	Result int    `json:"result"`
	Reason string `json:"reason"`
}

func Register(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	email := req.FormValue("email")
	secret := req.FormValue("secret")
	u, err := model.Models.User.Register(email, secret)
	if err != nil {
		response(w, err)
		return
	}
	response(w, &RegisterResp{
		Result:     200,
		Uid:        u.Id.Hex(),
		Token:      u.Token,
		MgrAddr:    NewMgrHost(req),
		NotifyAddr: NewNotifyHost(req),
	})
}

type LoginResp struct {
	Result     int      `json:"result"`
	Reason     string   `json:"reason,omitempty"`
	Token      string   `json:"token"`
	Uid        string   `json:"uid"`
	MgrAddr    []string `json:"mgraddr"`
	NotifyAddr []string `json:"notifyaddr"`
}

type InitResp struct {
	Result     int      `json:"result"`
	MgrAddr    []string `json:"mgraddr"`
	NotifyAddr []string `json:"notifyaddr"`
}

func Login(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	email := req.PostFormValue("email")
	secret := req.PostFormValue("secret")
	uid, token, err := model.Models.User.Login(email, secret)
	if err != nil {
		response(w, err)
		return
	}
	response(w, &LoginResp{
		Result:     200,
		Uid:        uid,
		Token:      token,
		MgrAddr:    NewMgrHost(req),
		NotifyAddr: NewNotifyHost(req),
	})
}

func Init(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	uid := req.PostFormValue("uid")
	token := req.PostFormValue("token")

	ok, err := model.Models.User.CheckToken(uid, token)
	if err != nil {
		response(w, err)
		return
	}
	if !ok {
		response(w, ErrInvalidToken)
		return
	}

	response(w, &InitResp{
		Result:     200,
		MgrAddr:    NewMgrHost(req),
		NotifyAddr: NewNotifyHost(req),
	})
}

func NewMgrHost(req *http.Request) []string {
	host := req.Host
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		host = host[:idx]
	}

	return []string{host + ":8682"}
}

func NewNotifyHost(req *http.Request) []string {
	host := req.Host
	if idx := strings.LastIndex(host, ":"); idx > 0 {
		host = host[:idx]
	}

	return []string{host + ":8683"}
}
