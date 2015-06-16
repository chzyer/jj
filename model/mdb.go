package model

import (
	"net/url"
	"time"

	"gopkg.in/mgo.v2"
)

type Mdb struct {
	session *mgo.Session
	DBName  string
}

func NewMdb(dbname string, mdb *mgo.Session) *Mdb {
	return &Mdb{session: mdb.Copy(), DBName: dbname}
}

func DialUrl(url_ string) (string, *mgo.Session, error) {
	u, _ := url.Parse("tcp://" + url_)

	session, err := mgo.DialWithTimeout(u.Host, time.Second)
	if err != nil {
		return "", nil, err
	}
	return u.Path[1:], session, nil
}

func (m *Mdb) Close() {
	m.session.Close()
}

func (m *Mdb) collection(c string) *mgo.Collection {
	return m.session.DB(m.DBName).C(c)
}

func (m *Mdb) One(c string, query M, result interface{}) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Find(query).One(result)
}

func (m *Mdb) Distinct(c string, key string, query M, result interface{}) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Find(query).Distinct(key, result)
}

func (m *Mdb) All(c string, query M, result interface{}) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Find(query).All(result)
}

func (m *Mdb) Count(c string, query M) (int, error) {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Find(query).Count()
}

func (m *Mdb) Insert(c string, docs ...interface{}) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Insert(docs...)
}

func (m *Mdb) Upsert(c string, selector, update M) error {
	session := m.session.Copy()
	defer session.Close()
	_, err := m.collection(c).Upsert(selector, update)
	return err
}

func (m *Mdb) Update(c string, selector, update M) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Update(selector, update)
}

func (m *Mdb) Remove(c string, selector M) error {
	session := m.session.Copy()
	defer session.Close()
	return m.collection(c).Remove(selector)
}
