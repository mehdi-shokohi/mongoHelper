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
)

var ErrorInsertFailed = errors.New("insert failed with no error")



type Transaction struct {
	connection *mongo.Client
	ctx        context.Context
}

func StartTransaction(Id string,ctx context.Context) Transaction {
	tr := Transaction{}
	tr.ctx =ctx
	tr.connection, _ = Holder.Read(Id)
	return tr
}

func (t *Transaction) EndTransaction(f func(sessionContext mongo.SessionContext) (result interface{}, err error)) (interface{}, error) {

	wc := writeconcern.New(writeconcern.W(1))
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	session, err := t.connection.StartSession()
	if err != nil {
		return nil, err
	}
	resp, err := session.WithTransaction(t.ctx, f, txnOpts)
	if err != nil {
		err = session.AbortTransaction(t.ctx)
	}

	defer session.EndSession(t.ctx)
	return resp, err

}

func (m *MongoContainer[T]) GetCollection() *mongo.Collection {
	// conn,_ := Holder.Read(id)
	// conn := mongoDB.GetConnection()
	var collection *mongo.Collection
	if cs, ok := m.Ctx.(mongo.SessionContext); ok {
		collection = cs.Client().Database(m.DatabaseName).Collection(m.CollectionName)
	} else {
		collection = m.Connection.Database(m.DatabaseName).Collection(m.CollectionName)

	}
	return collection
}

type MongoContainer[T any] struct {
	Model          T
	Ctx            context.Context
	Connection     *mongo.Client
	DatabaseName   string
	CollectionName string
}

// func (m *MongoContainer[T]) ConnectionManager(Op func(ctx context.Context, collection *mongo.Collection) (interface{}, error)) (interface{}, error) {
// 	return func() (interface{}, error) {
// 		var collection *mongo.Collection
// 		if cs, ok := m.Ctx.(mongo.SessionContext); ok {
// 			collection = cs.Client().Database(m.DatabaseName).Collection(m.CollectionName)
// 		} else {
// 			collection = m.Connection.Database(m.DatabaseName).Collection(m.CollectionName)

// 		}
// 		return Op(m.Ctx, collection)
// 	}()
// }

func (m *MongoContainer[T]) Insert() (result *mongo.InsertOneResult, err error) {

	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {

	insertResult, err := m.GetCollection().InsertOne(m.Ctx, m.Model)
	if err != nil {

		return nil, err
	}
	if insertResult == nil {
		return nil, ErrorInsertFailed
	}

	return insertResult, err
	// })
}

func (m *MongoContainer[T]) Update(newValue T, findCondition bson.D) (result *mongo.UpdateResult, err error) {

	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
	updateFilter := bson.D{{Key: "$set", Value: newValue}}
	updatedResult, err := m.GetCollection().UpdateOne(m.Ctx, findCondition, updateFilter)
	if err != nil {
		return nil, err

	}
	if updatedResult == nil || updatedResult.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return updatedResult, err
	// })
}
func (m *MongoContainer[T]) UpdateMany(filter interface{}, update interface{}, options ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
	updatedResult, err := m.GetCollection().UpdateMany(m.Ctx, filter, update, options...)
	if err != nil {
		return nil, err

	}
	if updatedResult == nil {
		return nil, mongo.ErrNoDocuments

	}
	return updatedResult, nil
	// })
}

func (m *MongoContainer[T]) FindOne(query *bson.D, options ...*options.FindOneOptions) (*T, error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
	one := m.GetCollection().FindOne(m.Ctx, query, options...)
	err := one.Decode(&m.Model)
	if err == nil {
		return &m.Model, nil
	}

	return nil, err
	// })
}

func (m *MongoContainer[T]) FindAll(query *bson.D, opts ...*options.FindOptions) ([]*T, error) {
	// var collection *mongo.Collection
	// if cs, ok := m.Ctx.(mongo.SessionContext); ok {
	// 	collection = cs.Client().Database(m.DatabaseName).Collection(m.CollectionName)
	// } else {
	// 	collection = m.Connection.Database(m.DatabaseName).Collection(m.CollectionName)

	// }
	collection := m.GetCollection()
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
			log.Println(err)
		}

		results = append(results, elem)

	}

	if err := cur.Close(m.Ctx); err != nil {
		return nil, err
	}
	return results, nil
}

// CountDocuments returns total documents founded by query.
func (m *MongoContainer[T]) CountDocuments(query *bson.D) (interface{}, error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
	return m.GetCollection().CountDocuments(m.Ctx, query)
	// })

}

// Aggregate as you know aggregate mongo pipeline!
func (m *MongoContainer[T]) Aggregate(pipeline mongo.Pipeline, opts ...*options.AggregateOptions) (result []T, err error) {
	// var collection *mongo.Collection
	// if cs, ok := m.Ctx.(mongo.SessionContext); ok {
	// 	collection = cs.Client().Database(m.DatabaseName).Collection(m.CollectionName)
	// } else {
	// 	collection = m.Connection.Database(m.DatabaseName).Collection(m.CollectionName)

	// }
	collection := m.GetCollection()
	aggResult, err := collection.Aggregate(m.Ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	results := make([]T, 0)

	for aggResult.Next(m.Ctx) {
		//Create a value into which the single document can be decoded
		elem := new(T)
		err := aggResult.Decode(elem)
		if err != nil {
			log.Println(err)
		}

		results = append(results, *elem)

	}

	if err := aggResult.Close(m.Ctx); err != nil {
		return nil, err
	}
	return results, nil

}

func (m *MongoContainer[T]) FindByID(id interface{}) (*T, error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
		one := m.GetCollection().FindOne(m.Ctx, &bson.D{{Key: "_id", Value: id}})
		err := one.Err()
		if err != nil {

			return nil, err
		}
		err = one.Decode(m.Model)
		return &m.Model, err
	// })
}

func (m *MongoContainer[T]) DeleteOne(b *bson.D, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
		delete, err := m.GetCollection().DeleteOne(m.Ctx, b, opts...)
		if err != nil {
			return nil, err
		}

		return delete, nil
	// })
}

func (m *MongoContainer[T]) DeleteMany(b *bson.D, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	// return m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
		delCount, err := m.GetCollection().DeleteMany(m.Ctx, b, opts...)
		if err != nil {
			return nil, err
		}

		return delCount, nil
	// })
}
