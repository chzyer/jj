package model

import (
	"regexp"

	"github.com/jj-io/jj/internal"

	"gopkg.in/logex.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	RegexpUserEmail     = regexp.MustCompile(`\w+@(\w+\.)+\w+`)
	ErrUserEmailInvalid = logex.Define("email '%v' is invalid")
	ErrUserPswdEmpty    = logex.Define("password is empty")
	ErrUserLoginFail    = logex.Define("incorrect username or password")

	ErrUserEmailAlreadyTaken = logex.Define("email '%v' is already taken")
)

type User struct {
	Id     bson.ObjectId `bson:"_id"`
	Email  string        `bson:"email"`
	Secret string        `bson:"secret"`
	Token  string        `bson:"token"`
	Valid  bool          `bson:"valid"`
}

func (u User) Index() []mgo.Index {
	return []mgo.Index{
		{Key: []string{"Email"}, Unique: true},
	}
}

type UserModel struct {
	*BaseModel
}

func NewUserModel(mdb *Mdb) *UserModel {
	return &UserModel{NewBaseModel(mdb, User{})}
}

func (um *UserModel) Register(email, secret string) (bson.ObjectId, error) {
	if !RegexpUserEmail.MatchString(email) {
		return "", ErrUserEmailInvalid.Format(email).SetCode(400)
	}
	if secret == "" {
		return "", logex.TraceError(ErrUserPswdEmpty).SetCode(400)
	}
	if taken, err := um.Find(email); err != nil {
		return "", logex.Trace(err)
	} else if taken {
		return "", ErrUserEmailAlreadyTaken.Format(email).SetCode(400)
	}
	u := &User{
		Id:     bson.NewObjectId(),
		Email:  email,
		Secret: secret,
		Token:  internal.GenUuid([]byte(secret)),
		Valid:  true,
	}

	if err := logex.Trace(um.Insert(u)); err != nil {
		return "", err
	}
	return u.Id, nil
}

func (um *UserModel) Find(email string) (bool, error) {
	if !RegexpUserEmail.MatchString(email) {
		return false, ErrUserEmailInvalid.Format(email).SetCode(400)
	}

	has, err := um.Has(M{"email": email})
	if err != nil {
		err = logex.Trace(err)
	}
	return has, err
}

func (um *UserModel) Login(email, secret string) (uid, token string, err error) {
	if !RegexpUserEmail.MatchString(email) {
		return "", "", ErrUserEmailInvalid.Format(email).SetCode(400)
	}
	if secret == "" {
		return "", "", ErrUserPswdEmpty.SetCode(400)
	}

	var u *User
	if err := um.One(M{
		"email":  email,
		"secret": secret,
	}, &u); err != nil {
		return "", "", logex.Trace(err)
	}
	if u == nil {
		return "", "", ErrUserLoginFail.SetCode(401)
	}
	return u.Id.Hex(), u.Token, nil
}
