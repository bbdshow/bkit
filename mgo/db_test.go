package mgo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	uri        = "mongodb://localhost:27017"
	database   = "mongo_test"
	collection = "test"
)

type doc struct {
	Name  string `bson:"name"`
	Value string `bson:"value"`
}

func TestDatabase_CreateIndexes(t *testing.T) {
	db, err := NewDatabase(context.Background(), DBConn{URI: uri, Database: database})
	if err != nil {
		t.Fatal(err)
	}

	manyIndex := []Index{
		{
			Collection: collection,
			Keys:       bson.D{{Key: "name", Value: 1}},
		},
		{
			Collection: collection,
			Name:       "name_value",
			Keys:       bson.D{{Key: "name", Value: 1}, {Key: "value", Value: -1}},
			Unique:     true,
		}}

	err = db.CreateIndexes(manyIndex)
	if err != nil {
		t.Fatal(err)
		return
	}

	cursor, err := db.Collection(collection).Indexes().List(nil)
	if err != nil {
		t.Fatal(err)
		return
	}
	var value interface{}
next:
	if cursor.Next(nil) {
		err = cursor.Decode(&value)
		if err != nil {
			t.Log(err) //return
		}

		if strings.Index(fmt.Sprintf("%#v", value), "name_value") != -1 {
			return
		}
		goto next
	}

	t.Fail()
}

func TestDatabase_Transaction(t *testing.T) {
	database := "tx_test"
	uri := "mongodb://root:111111@192.168.10.25:27017,192.168.10.26:27017/?authSource=admin$replicaSet=fbj&authSource=admin&maxPoolSize=50"
	db, err := NewDatabase(context.Background(), DBConn{URI: uri, Database: database})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Transaction(context.Background(), func(session SessionContext) error {
		type r struct {
			Name string    `bson:"name"`
			At   time.Time `bson:"at"`
		}
		iRet, err := db.Collection("tx").InsertOne(session, &r{Name: "tx_key", At: time.Now()})
		if err != nil {
			return err
		}
		log.Println(iRet)
		return fmt.Errorf("exception")
	})
	if err != nil {
		t.Fatal(err)
	}
}

type tableA struct {
	Name string `bson:"name"`
}

func (*tableA) TableName() string {
	return "table_a"
}

type tableB struct {
	Name string `bson:"name"`
}

func (tableB) TableName() string {
	return "table_b"
}
func TestDatabase_collectionName(t *testing.T) {
	tableA := &tableA{}
	tableB := tableB{}
	if CollectionName(tableA) != "table_a" {
		t.Fail()
	}
	if CollectionName(&tableB) != "table_b" {
		t.Fail()
	}
}
