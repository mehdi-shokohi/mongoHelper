package mongoHelper

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	conf "github.com/mehdi-shokohi/mongoHelper/config"
	db "github.com/mehdi-shokohi/mongoHelper/mongoPools"
)

var ErrorInsertFailed = errors.New("insert failed with no error")


type Transaction struct {
	WMongo     *db.MongoWriteDB
	Connection *mongo.Database
}

func StartTransaction() *Transaction {
	tr := new(Transaction)
	tr.WMongo = db.GetWriteDB()
	tr.Connection = tr.WMongo.GetConnection()
	return tr
}

func (t *Transaction) EndTransaction(f func(sessionContext mongo.SessionContext) (result interface{}, err error)) (interface{}, error) {

	wc := writeconcern.New(writeconcern.W(1))
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := t.Connection.Client().StartSession()
	if err != nil {
		return nil, err
	}
	resp, err := session.WithTransaction(context.Background(), f, txnOpts)
	if err != nil {
		err=session.AbortTransaction(context.Background())
	}

	defer session.EndSession(context.Background())
	defer t.WMongo.Release(t.Connection)
	return resp, err

}

func GetCollection(collectionName string) (*mongo.Collection, func()) {
	mongoDB := db.GetWriteDB()
	conn := mongoDB.GetConnection()
	return conn.Collection(collectionName), func() { mongoDB.Release(conn) }
}

type MongoContainer[T any] struct {
	Model          T
	Ctx            context.Context
	CollectionName string
}

func NewMongo[T any](ctx context.Context, colName string, model T) *MongoContainer[T] {
	m := new(MongoContainer[T])
	m.Model = model
	m.Ctx = ctx
	m.CollectionName = colName
	return m
}
func (m *MongoContainer[T]) Insert() (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}

	insertResult, err := collection.InsertOne(m.Ctx, m.Model)
	if err != nil {

		return nil, err
	}
	if insertResult == nil {
		return nil, ErrorInsertFailed
	}

	return insertResult, err
}


func (m *MongoContainer[T]) Update(newValue T, findCondition bson.D) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}

	updateFilter := bson.D{{Key: "$set", Value: newValue}}
	updatedResult, err := collection.UpdateOne(m.Ctx, findCondition, updateFilter)
	if err != nil {
		return nil, err

	}
	if updatedResult == nil || updatedResult.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
		return
	}
	return updatedResult, err
}
func (m *MongoContainer[T]) UpdateMany(filter interface{}, update interface{}, options ...*options.UpdateOptions) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}
	updatedResult, err := collection.UpdateMany(m.Ctx, filter, update, options...)
	if err != nil {
		return nil, err

	}
	if updatedResult == nil {
		return nil, mongo.ErrNoDocuments

	}
	return updatedResult, nil
}
func (m *MongoContainer[T]) FindOne(query *bson.D, options ...*options.FindOneOptions) (interface{}, error) {
	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}
	one := collection.FindOne(m.Ctx, query, options...)
	err := one.Decode(m.Model)
	if err == nil {
		return one, nil
	}

	return nil, err
}

func (m *MongoContainer[T]) FindAll(query *bson.D, opts ...*options.FindOptions) ([]*T, error) {
	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}
	results := make([]*T, 0)
	cur, err := collection.Find(m.Ctx, query, opts...)
	if err != nil {
		return nil, err
	}
	for cur.Next(m.Ctx) {
		//Create a value into which the single document can be decoded
		elem := new(T)
		err := cur.Decode(elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem)

	}
	cur.Close(m.Ctx)
	return results, nil
}

func (m *MongoContainer[T]) FindByID(id interface{}) (interface{}, error) {
	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}
	one := collection.FindOne(m.Ctx, &bson.D{{Key: "_id", Value: id}})
	err := one.Err()
	if err != nil {

		return nil, err
	}
	err = one.Decode(m.Model)
	return one, err
}

func (m *MongoContainer[T]) DeleteOne(b *bson.D, opts ...*options.DeleteOptions) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}

	delete, err := collection.DeleteOne(m.Ctx, b, opts...)
	if err != nil {
		return nil, err
	}

	return delete, nil
}

func (m *MongoContainer[T]) DeleteMany(b *bson.D, opts ...*options.DeleteOptions) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(m.CollectionName)
		defer release()
	}

	delCount, err := collection.DeleteMany(m.Ctx, b, opts...)
	if err != nil {
		return nil, err
	}

	return delCount, nil
}
