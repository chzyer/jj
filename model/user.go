package model

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id     bson.ObjectId `bson:"_id"`
	Email  string        `bson:"email"`
	Secret string        `bson:"secret"`
	Token  string        `bson:"token"`
}

type UserModel struct {
	*BaseModel
}

func NewUserModel(mdb *mgo.Session) *UserModel {
	return &UserModel{NewBaseModel(mdb, User{})}
}

func (u *UserModel) Register(email, secret string) error {
	return nil
}

func (u *UserModel) Login(email, secret string) (token string, err error) {
	return "", nil
}
