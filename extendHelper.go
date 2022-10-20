package main

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mehdi-shokohi/mongoHelper/mongoHelper"
)
import "go.mongodb.org/mongo-driver/bson/primitive"

// User Database Model
type Model struct {
	ID        *primitive.ObjectID `json:"-"   bson:"_id,omitempty"`
	FirstName string              `json:"first_name"   bson:"first_name"`
	LastName  string              `json:"last_name"   bson:"last_name"`
	Username  string              `json:"username"   bson:"username"`
	Password  string              `json:"password"   bson:"password"`
	Admin     bool                `json:"admin"   bson:"admin"`
}
type MongoHelper[T any] struct {
	mongoHelper.MongoContainer[T]
}

func (m *MongoHelper[T]) MyFindOne(query *bson.D, options ...*options.FindOneOptions) (*Model, error) {
	user, err := m.ConnectionManager(func(ctx context.Context, collection *mongo.Collection) (interface{}, error) {
		one := collection.FindOne(ctx, query, options...)
		err := one.Decode(m.Model)
		if err == nil {
			return m.Model, nil
		}
		return nil, err
	})
	if err != nil {
		return nil, err
	}
	return user.(*Model), nil
}

func MyNewMongo[T any](ctx context.Context, colName string, model T) *MongoHelper[T] {
	m := new(MongoHelper[T])
	m.Model = model
	m.Ctx = ctx
	m.CollectionName = colName
	return m
}
