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
}

func response(w http.ResponseWriter, obj interface{}) {
	switch result := obj.(type) {
	case error:
		logex.Error(result)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(500)
		io.WriteString(w, result.Error())
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
	response(w, url.Values{
		"id": {id.Hex()},
	})
}

func Login(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	email := req.FormValue("email")
	secret := req.FormValue("secret")
	token, err := model.Models.User.Login(email, secret)
	if err != nil {
		response(w, err)
		return
	}
	response(w, url.Values{
		"token": {token},
	})
}
