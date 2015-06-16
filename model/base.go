package model

import (
	"reflect"

	"gopkg.in/logex.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Models *models

var (
	ErrProfileEmpty = logex.Define("profile is empty")
)

func Init(mdb *mgo.Session) {
	Models = newModels(mdb)
}

func IsPanicError(err error) bool {
	if err == nil {
		return false
	}
	return !logex.Equal(err, mgo.ErrNotFound)
}

type models struct {
	User *UserModel
}

func newModels(mdb *mgo.Session) *models {
	return &models{}
}

type M bson.M

type models struct {
	User *UserModel
}

type BaseModel struct {
	ins  reflect.Type
	Name string
	mdb  *mgo.Session
}

func NewBaseModel(mdb *mgo.Session, ins interface{}) *BaseModel {
	t := reflect.TypeOf(ins)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return &BaseModel{
		ins:  t,
		Name: t.Name(),
		mdb:  mdb,
	}
}

func (b *BaseModel) One(query, result interface{}) error {
	session := b.mdb.Copy()
	session.Close()
	return session.One(b.Name, query, result)
}

func (b *BaseModel) All(query, result interface{}) error {
	return b.mdb.All(b.Name, query, result)
}

func (b *BaseModel) Distinct(key string, query, result interface{}) error {
	return b.mdb.WithC(b.Name, func(c *mgo.Collection) error {
		return c.Find(query).Distinct(key, result)
	})
}

func (b *BaseModel) Count(query interface{}) (n int) {
	return b.mdb.Count(b.Name, query)
}

func (b *BaseModel) Insert(data ...interface{}) error {
	return b.mdb.Insert(b.Name, data...)
}

func (b *BaseModel) Update(selector, change interface{}) error {
	return b.mdb.Update(b.Name, selector, change)
}

func (b *BaseModel) Upsert(selector, change interface{}) error {
	return b.mdb.Upsert(b.Name, selector, change)
}

func (b *BaseModel) Remove(selector interface{}) error {
	return b.mdb.Remove(b.Name, selector)
}
