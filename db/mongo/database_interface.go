package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type DataBaseInterface interface {
	// extend quick function
	Exists(ctx context.Context, collection string, filter interface{}) (bool, error)
	FindOne(ctx context.Context, collection string, filter interface{}, doc interface{}, opt ...*options.FindOneOptions) (bool, error)
	Find(ctx context.Context, collection string, filter bson.M, docs interface{}, opt ...*options.FindOptions) error
	FindCount(ctx context.Context, collection string, filter interface{}, docs interface{}, findOpt *options.FindOptions, countOpt *options.CountOptions) (int64, error)
	ListCollectionNames(ctx context.Context, prefix ...string) ([]string, error)
	UpsertCollectionIndexMany(indexMany ...[]Index) error

	// 原有方法
	Client() *mongo.Client
	Name() string
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
	Aggregate(ctx context.Context, pipeline interface{},
		opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) *mongo.SingleResult
	RunCommandCursor(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (*mongo.Cursor, error)
	Drop(ctx context.Context) error
	ListCollectionSpecifications(ctx context.Context, filter interface{},
		opts ...*options.ListCollectionsOptions) ([]*mongo.CollectionSpecification, error)
	ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (*mongo.Cursor, error)
	ReadConcern() *readconcern.ReadConcern
	ReadPreference() *readpref.ReadPref
	WriteConcern() *writeconcern.WriteConcern
	Watch(ctx context.Context, pipeline interface{},
		opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
	CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error
	CreateView(ctx context.Context, viewName, viewOn string, pipeline interface{},
		opts ...*options.CreateViewOptions) error
}
