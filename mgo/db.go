package mgo

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func NewDatabase(ctx context.Context, conn DBConn) (*Database, error) {
	cli, err := NewMongoClientByURI(ctx, conn.URI)
	if err != nil {
		return nil, err
	}
	return &Database{Database: cli.Database(conn.Database)}, nil
}

// CollectionName 获取集合名称 支持 string, *mongo.Collection, struct TableName() string
func CollectionName(collection interface{}) string {
	switch v := collection.(type) {
	case string:
		return v
	case *mongo.Collection:
		return v.Name()
	default:
		// 反射 collection 是否为结构体， 是否存在方法 TableName() string
		value := reflect.ValueOf(collection)
		if value.Kind() == reflect.Ptr {
			// 是否为结构体
			if value.Elem().Kind() == reflect.Struct {
				// 是否存在方法 TableName() string
				if method := value.MethodByName("TableName"); method.IsValid() {
					if results := method.Call(nil); len(results) > 0 {
						if name, ok := results[0].Interface().(string); ok {
							return name
						}
					}
				}
			}
		}
	}
	return ""
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

// Transaction 只支持Mongodb实例为 副本集或者分片集群部署
//
//	tx := func(session mongo.SessionContext) error {
//		collection := db.Database().Collection("testcollection")
//		_, err := collection.InsertOne(session, bson.M{"name": "Alice", "age": 30})
//		return err
//	}
func (db *Database) Transaction(ctx context.Context, tx func(session SessionContext) error) error {
	return db.Client().UseSession(ctx, func(session mongo.SessionContext) error {
		err := session.StartTransaction(options.Transaction().
			SetReadConcern(readconcern.Snapshot()).
			SetWriteConcern(writeconcern.Majority()))
		if err != nil {
			return err
		}

		if err := tx(session); err != nil {
			e := session.AbortTransaction(ctx)
			if e != nil {
				return fmt.Errorf("tx: %v abort: %v", err, e)
			}
			return err
		}

		for {
			err := session.CommitTransaction(ctx)
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
