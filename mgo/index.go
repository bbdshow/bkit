package mgo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Index struct {
	Collection         string
	Name               string // index name
	Keys               bson.D
	Unique             bool  // unique index
	ExpireAfterSeconds int32 // ttl index
}

func (i Index) Validate() error {
	if i.Collection == "" {
		return errors.New("collection required")
	}
	if len(i.Keys) == 0 {
		return errors.New("keys required")
	}
	return nil
}

var Indexes []Index

// CreateIndex add indexes to Indexes, waiting create
func CreateIndex(indexes []Index) {
	Indexes = append(Indexes, indexes...)
}

func (db *Database) indexesToModel(indexes []Index) map[string][]mongo.IndexModel {
	indexModels := make(map[string][]mongo.IndexModel)
	for _, index := range indexes {
		if err := index.Validate(); err != nil {
			panic(err)
		}
		model := mongo.IndexModel{
			Keys: index.Keys,
		}
		opt := options.Index()
		if index.Name != "" {
			opt.SetName(index.Name)
		}
		opt.SetUnique(index.Unique)

		if index.ExpireAfterSeconds > 0 {
			opt.SetExpireAfterSeconds(index.ExpireAfterSeconds)
		}

		model.Options = opt

		v, ok := indexModels[index.Collection]
		if ok {
			indexModels[index.Collection] = append(v, model)
		} else {
			indexModels[index.Collection] = []mongo.IndexModel{model}
		}
	}
	return indexModels
}

// CreateIndexes 创建索引，已经存在的索引不会重复创建
func (db *Database) CreateIndexes(indexes []Index) error {
	indexModels := db.indexesToModel(indexes)
	for collection, index := range indexModels {
		_, err := db.Collection(collection).Indexes().CreateMany(context.Background(), index)
		if err != nil {
			return fmt.Errorf("collection: %s %v", collection, err)
		}
	}
	return nil
}
