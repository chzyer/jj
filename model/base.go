package model

import (
	"reflect"
	"sync"

	"gopkg.in/logex.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Models *models
var once sync.Once

var (
	ErrProfileEmpty  = logex.Define("profile is empty")
	ErrInvalidObject = logex.Define("Invalid input to ObjectIdHex")
)

func Init(url_ string) (err error) {
	once.Do(func() {
		var dbName string
		var mdb *mgo.Session
		dbName, mdb, err = DialUrl(url_)
		if err != nil {
			return
		}
		Models = newModels(dbName, mdb)
	})
	return logex.Trace(err)
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

func newModels(dbName string, session *mgo.Session) *models {
	return &models{
		User: NewUserModel(NewMdb(dbName, session)),
	}
}

type M bson.M

func BsonObjectId(id string) (bson.ObjectId, error) {
	if !bson.IsObjectIdHex(id) {
		return "", ErrInvalidObject.Trace()
	}
	return bson.ObjectIdHex(id), nil
}

type Indexer interface {
	Index() []mgo.Index
}

type BaseModel struct {
	ins  reflect.Type
	Name string
	mdb  *Mdb
}

func NewBaseModel(mdb *Mdb, ins interface{}) *BaseModel {
	t := reflect.TypeOf(ins)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	b := &BaseModel{
		ins:  t,
		Name: t.Name(),
		mdb:  mdb,
	}
	b.ensureIndex(ins)
	return b
}

func (b *BaseModel) ensureIndex(obj interface{}) (err error) {
	if i, ok := obj.(Indexer); ok {
		for _, index := range i.Index() {
			err = b.mdb.collection(b.Name).EnsureIndex(index)
			if err != nil {
				return logex.Trace(err)
			}
		}
	}
	return nil
}

func (b *BaseModel) One(query M, result interface{}) error {
	err := b.mdb.One(b.Name, query, result)
	if err == mgo.ErrNotFound {
		err = nil
	}
	return err
}

func (b *BaseModel) All(query M, result interface{}) error {
	return b.mdb.All(b.Name, query, result)
}

func (b *BaseModel) Distinct(key string, query M, result interface{}) error {
	return b.mdb.Distinct(b.Name, key, query, result)
}

func (b *BaseModel) Has(query M) (bool, error) {
	var o interface{}
	err := b.mdb.One(b.Name, query, &o)
	if err != nil {
		if err == mgo.ErrNotFound {
			err = nil
		}
		return false, err
	}
	return o != nil, nil
}

func (b *BaseModel) Count(query M) (n int, err error) {
	return b.mdb.Count(b.Name, query)
}

func (b *BaseModel) Insert(data ...interface{}) error {
	return b.mdb.Insert(b.Name, data...)
}

func (b *BaseModel) Update(selector, change M) error {
	return b.mdb.Update(b.Name, selector, change)
}

func (b *BaseModel) Upsert(selector, change M) error {
	return b.mdb.Upsert(b.Name, selector, change)
}

func (b *BaseModel) Remove(selector M) error {
	return b.mdb.Remove(b.Name, selector)
}
