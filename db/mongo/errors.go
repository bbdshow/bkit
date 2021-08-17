package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
)

var (
	ErrNotMatched       = errors.New("not matched")
	ErrInstanceNotFound = errors.New("mongodb instance db not found")
	ErrNoDocuments      = mongo.ErrNoDocuments
)

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "E11000 duplicate key")
}
