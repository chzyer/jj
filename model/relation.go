package model

import (
	"time"

	"gopkg.in/mgo.v2"
)

type Relation struct {
	Uids     [2]string `bson:"uids"`
	Approved [2]bool   `bson:"approved"`
	Source   string    `bson:"source"`
	Date     time.Time `bson:"date"`
}

func (r Relation) Index() []mgo.Index {
	return []mgo.Index{
		{Key: []string{"uids"}, Unique: true},
	}
}

type RelationModel struct {
	*BaseModel
}

func NewRelationModel(mdb *Mdb) *RelationModel {
	return &RelationModel{NewBaseModel(mdb, User{})}
}

func (r *RelationModel) Add() {

}
