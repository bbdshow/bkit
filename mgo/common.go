package mgo

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotMatched       = errors.New("not matched")
	ErrDatabaseNotFound = errors.New("mongodb database not found")
	ErrNoDocuments      = mongo.ErrNoDocuments
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "E11000 duplicate key")
}

var Indexes []Index

// CreateIndex add indexes to Indexes, waiting create
func CreateIndex(indexes []Index) {
	Indexes = append(Indexes, indexes...)
}
