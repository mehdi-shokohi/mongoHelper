package mongoHelper

import (
	"context"
	"errors"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	db "github.com/mehdi-shokohi/mongoHelper/mongoPools"
)
import conf "github.com/mehdi-shokohi/mongoHelper/config"

var ErrorInsertFailed = errors.New("insert failed with no error")

type DecoderMap func(m map[string]interface{}) error

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
		//err=session.AbortTransaction(context.Background())
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

type MongoContainer[T Recorder] struct {
	Model Recorder
	Ctx   context.Context
}

func NewMongo[T Recorder](ctx context.Context, model T) *MongoContainer[T] {
	m := new(MongoContainer[T])
	m.Model = model
	m.Ctx = ctx
	return m
}
func (m *MongoContainer[T]) Insert() (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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

// func(m *MongoContainer[T]) Update(ctx context.Context, record T)  {
// 	res := make(chan error)
// 	go Update(ctx, record, res)
// 	return res

// }
func UpdateMapGo(ctx context.Context, collectionName string, id *primitive.ObjectID, newData map[string]interface{}) chan error {
	res := make(chan error)
	go func() {
		var collection *mongo.Collection
		if cs, ok := ctx.(mongo.SessionContext); ok {
			collection = cs.Client().Database(conf.GetMongodbName()).Collection(collectionName)
		} else {
			var release func()
			collection, release = GetCollection(collectionName)
			defer release()
		}

		updateFilter := bson.D{{Key: "$set", Value: newData}}
		updatedResult, err := collection.UpdateOne(ctx, bson.D{{"_id", id}}, updateFilter)
		if err != nil {
			res <- err
			return
		}
		if updatedResult == nil || updatedResult.MatchedCount == 0 {
			res <- mongo.ErrNoDocuments
			return
		}
		res <- nil
	}()
	return res
}
func (m *MongoContainer[T]) Update(newValue T, findCondition bson.D) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
		defer release()
	}
	one := collection.FindOne(m.Ctx, query, options...)
	err := one.Decode(m.Model)
	if err == nil {
		return one, nil
	}

	return nil, err
}

type Decoder func(model Recorder) error

func (m *MongoContainer[T]) FindAll(query *bson.D, opts ...*options.FindOptions) ([]*T, error) {
	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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
		err := cur.Decode(&elem)
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
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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

func (m *MongoContainer[T]) DeleteOneGo(b *bson.D, opts ...*options.DeleteOptions) (result interface{}, err error) {

	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
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
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(m.Model.GetCollectionName())
	} else {
		var release func()
		collection, release = GetCollection(m.Model.GetCollectionName())
		defer release()
	}

	delCount, err := collection.DeleteMany(m.Ctx, b, opts...)
	if err != nil {
		return nil, err
	}

	return delCount, nil
}

type AdvancedQuery struct {
	_collection string
	_pipeline   []bson.M
	_limit      bson.M
	_skip       bson.M
	_c          context.Context
}

func (aq *AdvancedQuery) QueryGo(ctx context.Context) chan Decoder {
	res := make(chan Decoder)
	go func() {
		cur, err := aq.Query(ctx)
		if err != nil {
			res <- func(_ Recorder) error {
				return err
			}
			close(res)
			return
		}
		for cur.Next(ctx) {
			res <- func(cursor mongo.Cursor) Decoder {
				return func(model Recorder) error {
					err := cursor.Decode(model)
					if err != nil {
						println(err.Error())
						return err
					}
					// model.SetIsDocumented(true)
					return nil
				}
			}(*cur)
		}
		close(res)
	}()
	return res
}
func (aq *AdvancedQuery) CountGo(ctx context.Context, count *int64) chan error {
	res := make(chan error)
	go func() {
		countRes, err := aq.Count(ctx)
		if err != nil {
			res <- err
		} else {
			*count = (countRes)
			res <- nil
		}
	}()
	return res
}
func (aq *AdvancedQuery) Query(ctx context.Context) (*mongo.Cursor, error) {
	var collection *mongo.Collection
	if cs, ok := ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(aq._collection)
	} else {
		var release func()
		collection, release = GetCollection(aq._collection)
		defer release()
	}
	pipe := aq._pipeline
	pipe = append(pipe, aq._limit)
	pipe = append(pipe, aq._skip)
	return collection.Aggregate(ctx, pipe)
}
func (aq *AdvancedQuery) Count(ctx context.Context) (int64, error) {
	var collection *mongo.Collection
	if cs, ok := ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(aq._collection)
	} else {
		var release func()
		collection, release = GetCollection(aq._collection)
		defer release()
	}
	pipe := aq._pipeline
	pipe = append(pipe, bson.M{"$count": "_aq_count"})

	return collection.CountDocuments(ctx, pipe)
}

func AdvanceQueryCursor(ctx context.Context, collectionName string, fields string, query string, paginationOptions ...int) (*mongo.Cursor, error) {
	var collection *mongo.Collection
	if cs, ok := ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(collectionName)
	} else {
		var release func()
		collection, release = GetCollection(collectionName)
		defer release()
	}
	queryParams := strings.Split(query, " ")
	fieldArray := strings.Split(fields, " ")
	signedFieldsArray := make([]bson.M, len(fieldArray))
	for i, f := range fieldArray {
		signedFieldsArray[i] = bson.M{"$toString": "$" + f}
	}

	d := make([]interface{}, 2)
	d[0] = bson.M{"$toLower": "$" + fieldArray[0]}
	d[1] = queryParams[0]

	addFields := bson.M{
		"_aq_selectedFields": bson.M{"$concat": signedFieldsArray},
		"_aq_score":          bson.M{"$indexOfCP": d},
	}

	matchConditions := make([]bson.M, len(queryParams))
	for i, q := range queryParams {
		matchConditions[i] = bson.M{"_aq_selectedFields": primitive.Regex{
			Pattern: "" + q + "",
			Options: "gi",
		}}
	}

	match := bson.M{"$and": matchConditions}

	limit := 100
	offset := 0
	if len(paginationOptions) > 0 {
		limit = paginationOptions[0]
	}
	if len(paginationOptions) > 1 {
		offset = paginationOptions[1]
	}

	return collection.Aggregate(ctx, []bson.M{
		{"$addFields": addFields},
		{"$match": match},
		{"$sort": bson.M{"_aq_score": 1}},
		{"$limit": limit},
		{"$skip": offset},
	})

}

func AdvanceQueryG(ctx context.Context, collectionName string, fields string, query string, paginationOptions ...int) chan Decoder {
	res := make(chan Decoder)
	go func() {
		cur, err := AdvanceQueryCursor(ctx, collectionName, fields, query, paginationOptions...)
		if err != nil {
			res <- func(_ Recorder) error {
				return err
			}
			close(res)
			return
		}
		for cur.Next(ctx) {
			res <- func(cursor mongo.Cursor) Decoder {
				return func(model Recorder) error {
					err := cursor.Decode(model)
					if err != nil {
						return err
					}
					// model.SetIsDocumented(true)
					return nil
				}
			}(*cur)
		}
		close(res)
	}()
	return res
}

func Count(ctx context.Context, CollectionName string, query interface{}, count *int64, res chan error, opts ...*options.CountOptions) {
	var collection *mongo.Collection
	if cs, ok := ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(conf.GetMongodbName()).Collection(CollectionName)
	} else {
		var release func()
		collection, release = GetCollection(CollectionName)
		defer release()
	}
	documents, err := collection.CountDocuments(ctx, query, opts...)
	if err != nil {
		res <- err
		return
	}
	*count = documents
	res <- nil
}

func CountGo(c context.Context, CollectionName string, query interface{}, count *int64, opts ...*options.CountOptions) chan error {
	res := make(chan error)
	go Count(c, CollectionName, query, count, res, opts...)
	return res
}
func CountCollectionDocGo(ctx context.Context, CollectionName string, query interface{}, count *int64, opts ...*options.CountOptions) chan error {
	res := make(chan error)
	go func() {
		mongoDB := db.GetWriteDB()
		dbase := mongoDB.GetConnection()
		result := dbase.RunCommand(ctx, bson.M{"collStats": CollectionName})
		var document bson.M
		err := result.Decode(&document)
		if err != nil {
			res <- err
			return
		}
		*count = int64(document["count"].(int32))
		res <- nil

	}()
	//go Count(c, CollectionName, query, count, res, opts...)
	return res
}
func CountSync(c context.Context, CollectionName string, query interface{}, count *int64, opts ...*options.CountOptions) error {
	res := make(chan error)
	go Count(c, CollectionName, query, count, res, opts...)
	return <-res
}

func AggregateGo(ctx context.Context, collectionName string, pipe mongo.Pipeline, opts ...*options.AggregateOptions) chan Decoder {
	result := make(chan Decoder)
	go func() {

		var collection *mongo.Collection
		if cs, ok := ctx.(mongo.SessionContext); ok {
			collection = cs.Client().Database(conf.GetMongodbName()).Collection(collectionName)
		} else {
			var release func()
			collection, release = GetCollection(collectionName)
			defer release()
		}
		cur, err := collection.Aggregate(ctx, pipe, opts...)
		if err != nil {
			result <- func(_ Recorder) error { return err }
			close(result)
			return
		}

		defer func() { _ = cur.Close(ctx) }()
		for cur.Next(ctx) {
			result <- func(cursor mongo.Cursor) Decoder {
				return func(model Recorder) error {
					err := cursor.Decode(model)
					if err != nil {
						return err
					}
					// model.SetIsDocumented(true)
					return nil
				}
			}(*cur)
		}
		close(result)

	}()
	return result
}
func AggregateMapGo(ctx context.Context, collectionName string, pipe mongo.Pipeline, opts ...*options.AggregateOptions) chan DecoderMap {
	result := make(chan DecoderMap)
	go func() {

		var collection *mongo.Collection
		if cs, ok := ctx.(mongo.SessionContext); ok {
			collection = cs.Client().Database(conf.GetMongodbName()).Collection(collectionName)
		} else {
			var release func()
			collection, release = GetCollection(collectionName)
			defer release()
		}
		cur, err := collection.Aggregate(ctx, pipe, opts...)
		if err != nil {
			result <- func(_ map[string]interface{}) error { return err }
			close(result)
			return
		}

		defer func() { _ = cur.Close(ctx) }()
		for cur.Next(ctx) {
			result <- func(cursor mongo.Cursor) DecoderMap {
				return func(model map[string]interface{}) error {
					err := cursor.Decode(model)
					if err != nil {
						return err
					}
					return nil
				}
			}(*cur)
		}
		close(result)

	}()
	return result
}
