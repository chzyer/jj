package model

import (
	"time"

	"gopkg.in/logex.v1"
	"gopkg.in/mgo.v2/bson"
)

var (
	ErrSessionInvalidId = logex.Define("invalid session id")
)

type Session struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	UserId   string        `bson:"userid" json:"userid"`
	ToUid    string        `bson:"touid" json:"touid"`
	LastDate time.Time     `bson:"lastDate" json:"lastDate"`
}

type SessionModel struct {
	*BaseModel
}

func NewSessionModel(mdb *Mdb) *SessionModel {
	return &SessionModel{NewBaseModel(mdb, Session{})}
}

func (sm *SessionModel) New(uid, toUid string) (s *Session, err error) {
	if !bson.IsObjectIdHex(uid) || !bson.IsObjectIdHex(toUid) {
		return nil, ErrUserInvalidUid.Trace()
	}

	err = logex.Trace(sm.One(M{
		"uid":   uid,
		"touid": toUid,
	}, &s))
	if err != nil {
		return nil, logex.Trace(err)
	}

	if s != nil {
		return
	}

	if err := logex.Trace(sm.Upsert(M{
		"uid":   uid,
		"touid": toUid,
	}, M{"$setOnInsert": M{
		"": "",
	}})); err != nil {
		return
	}
}

func (sm *SessionModel) Get(sessionId string) (s *Session, err error) {
	id, ok := BsonObjectId(sessionId)
	if !ok {
		return nil, ErrSessionInvalidId.Trace()
	}
	err = logex.Trace(sm.One(M{
		"_id": id,
	}, &s))
	return
}

func (s *SessionModel) GetList(uid string) (list []*Session, err error) {
	if !bson.IsObjectIdHex(uid) {
		return nil, ErrUserInvalidUid.Trace()
	}

	err = logex.Trace(s.All(M{"userid": uid}, &list))
	return
}
