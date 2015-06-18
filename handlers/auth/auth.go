package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/jj-io/jj/model"

	"gopkg.in/logex.v1"
)

func InitHandler(mux *http.ServeMux) {
	mux.HandleFunc("/auth/register", Register)
	mux.HandleFunc("/auth/login", Login)
	mux.HandleFunc("/auth/init", Init)
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
	Result int    `json:"result"`
	Uid    string `json:"uid"`
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
	id, err := model.Models.User.Register(email, secret)
	if err != nil {
		response(w, err)
		return
	}
	response(w, &RegisterResp{
		Result: 200,
		Uid:    id.Hex(),
	})
}

type LoginResp struct {
	Result int    `json:"result"`
	Token  string `json:"token"`
	Uid    string `json:"uid"`
	*InitResp
}

type InitResp struct {
	NotifyServerAddr []string `json:"notify_server_addr"`
	OpServerAddr     []string `json:"op_server_addr"`
}

func Login(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	email := req.FormValue("email")
	secret := req.FormValue("secret")
	uid, token, err := model.Models.User.Login(email, secret)
	if err != nil {
		response(w, err)
		return
	}
	response(w, &LoginResp{
		Result: 200,
		Uid:    uid,
		Token:  token,
	})
}

func Init(w http.ResponseWriter, req *http.Request) {
	// verify

}
