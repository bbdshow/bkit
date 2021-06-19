package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoClientByURI uri = mongo://...
func NewMongoClientByURI(ctx context.Context, uri string) (*mongo.Client, error) {
	opt := options.Client().ApplyURI(uri)
	return NewMongoClient(ctx, opt)
}

// NewMongoClient client and ping
func NewMongoClient(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	client, err := mongo.NewClient(opts...)
	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
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
	defer cursor.Close(ctx)

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
	defer cursor.Close(ctx)
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

type Index struct {
	Collection         string
	Name               string // 指定索引名称
	Keys               bson.D
	Unique             bool  // 唯一索引
	Background         bool  // 非阻塞创建索引
	ExpireAfterSeconds int32 // 多少秒后过期
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
			opt.SetBackground(index.Background)

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
