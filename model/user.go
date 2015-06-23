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
	ErrUserInvalidUid   = logex.Define("invalid uid")
	ErrUserIdNotFound   = logex.Define("user not found")

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
		{Key: []string{"email"}, Unique: true},
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
		Token:  internal.GenUserToken(),
		Valid:  true,
	}

	if err := logex.Trace(um.Insert(u)); err != nil {
		return "", err
	}
	return u.Id, nil
}

func (um *UserModel) CheckToken(uid, token string) (bool, error) {
	id, ok := BsonObjectId(uid)
	if !ok {
		return false, ErrUserInvalidUid.Trace()
	}
	has, err := um.Has(M{
		"_id":   id,
		"token": token,
	})
	if err != nil {
		return false, logex.Trace(err)
	}
	return has, nil
}

func (um *UserModel) HasUid(uid string) (bool, error) {
	if !bson.IsObjectIdHex(uid) {
		return false, ErrUserInvalidUid.Trace()
	}
	has, err := um.Has(M{"uid": uid})
	if err != nil {
		return false, logex.Trace(err)
	}
	return has, nil
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

func (um *UserModel) GetToken(uid string) (token string, err error) {
	id, ok := BsonObjectId(uid)
	if !ok {
		return "", ErrUserInvalidUid.Trace()
	}

	var u *User
	if err := um.One(M{
		"_id": id,
	}, &u); err != nil {
		return "", logex.Trace(err)
	}
	return u.Token, nil
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
