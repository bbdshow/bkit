package mgo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DataBaseInterface interface {
	// 自定义快捷方法

	Exists(ctx context.Context, collection string, filter interface{}) (bool, error)
	FindOne(ctx context.Context, collection string, filter interface{}, doc interface{}, opt ...*options.FindOneOptions) (bool, error)
	Find(ctx context.Context, collection string, filter bson.M, docs interface{}, opt ...*options.FindOptions) error
	FindCount(ctx context.Context, collection string, filter interface{}, docs interface{}, findOpt *options.FindOptions, countOpt *options.CountOptions) (int64, error)
	ListCollectionNames(ctx context.Context, prefix ...string) ([]string, error)
	UpsertCollectionIndexMany(indexMany ...[]Index) error
	Transaction(ctx context.Context, tx func(session SessionContext) error) error

	// 原始方法

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

type SessionContext mongo.SessionContext

// NewMongoClientByURI uri = mongo://...
func NewMongoClientByURI(ctx context.Context, uri string) (*mongo.Client, error) {
	opt := options.Client().ApplyURI(uri)
	return NewMongoClient(ctx, opt)
}

// NewMongoClient client and ping
func NewMongoClient(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}
	// ping
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Ping(ctxTimeout, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// URIToHosts parse to hosts
func URIToHosts(uri string) []string {
	opt := options.Client().ApplyURI(uri)
	return opt.Hosts
}

// Database packaging some method, shortcut
type Database struct {
	*mongo.Database
}

func NewDatabase(ctx context.Context, uri, database string) (*Database, error) {
	cli, err := NewMongoClientByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return &Database{Database: cli.Database(database)}, nil
}

func (db *Database) Exists(ctx context.Context, collection string, filter interface{}) (bool, error) {
	opt := options.Count()
	opt.SetLimit(1)
	c, err := db.Database.Collection(collection).CountDocuments(ctx, filter, opt)
	return c > 0, err
}

func (db *Database) FindOne(ctx context.Context, collection string, filter interface{}, doc interface{}, opt ...*options.FindOneOptions) (bool, error) {
	err := db.Database.Collection(collection).FindOne(ctx, filter, opt...).Decode(doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (db *Database) Find(ctx context.Context, collection string, filter bson.M, docs interface{}, opt ...*options.FindOptions) error {
	cursor, err := db.Database.Collection(collection).Find(ctx, filter, opt...)
	if err != nil {
		return err
	}
	if err := cursor.All(ctx, docs); err != nil {
		return err
	}
	return nil
}

func (db *Database) FindCount(ctx context.Context, collection string, filter interface{}, docs interface{}, findOpt *options.FindOptions, countOpt *options.CountOptions) (int64, error) {
	c, err := db.Database.Collection(collection).CountDocuments(ctx, filter, countOpt)
	if err != nil {
		return 0, err
	}
	cursor, err := db.Database.Collection(collection).Find(ctx, filter, findOpt)
	if err != nil {
		return 0, err
	}
	if err := cursor.All(ctx, docs); err != nil {
		return 0, err
	}
	return c, nil
}

func (db *Database) ListCollectionNames(ctx context.Context, prefix ...string) ([]string, error) {
	filter := bson.M{}
	if len(prefix) > 0 && prefix[0] != "" {
		filter["name"] = primitive.Regex{
			Pattern: prefix[0],
		}
	}
	return db.Database.ListCollectionNames(ctx, filter)
}

// Transaction Only supports single db
func (db *Database) Transaction(ctx context.Context, tx func(session SessionContext) error) error {
	return db.Client().UseSession(ctx, func(session mongo.SessionContext) error {
		err := session.StartTransaction(options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.Majority()))
		if err != nil {
			return err
		}

		if err := tx(session); err != nil {
			e := session.AbortTransaction(session)
			if e != nil {
				return fmt.Errorf("tx: %v abort: %v", err, e)
			}
			return err
		}

		for {
			err := session.CommitTransaction(session)
			switch e := err.(type) {
			case nil:
				return nil
			case mongo.CommandError:
				if e.HasErrorLabel("UnknownTransactionCommitResult") {
					continue
				}
				return e
			default:
				return e
			}
		}
	})
}

type Index struct {
	Collection         string
	Name               string // index name
	Keys               bson.D
	Unique             bool  // unique index
	Background         bool  // background create index
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

func (db *Database) UpsertCollectionIndexMany(indexMany ...[]Index) error {
	indexModels := make(map[string][]mongo.IndexModel)
	for _, many := range indexMany {
		for _, index := range many {
			if err := index.Validate(); err != nil {
				return err
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

		for collection, index := range indexModels {
			_, err := db.Collection(collection).Indexes().CreateMany(context.Background(), index)
			if err != nil {
				return fmt.Errorf("collection: %s %v", collection, err)
			}
		}
	}

	return nil
}
