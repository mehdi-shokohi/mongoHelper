package mongoHelper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       *primitive.ObjectID `json:"id,omitempty"   bson:"_id,omitempty"`
	Password string              `json:"password,omitempty"   bson:"password"`
	Username string              `json:"username"   bson:"userName"`
	Status   string              `json:"status" bson:"status"`
}

const uri = "mongodb://localhost:27018/?readPreference=primary&appname=GOmongoHelper&directConnection=true&ssl=false"
const dbId = "mongo_rw_partition_1"
const database = "contents"


func NewMongo[T any](ctx context.Context,collection string , model T)MongoContainer[T]{
	db:=NewMongoById(ctx,dbId,uri,database,collection,model)
	return db
}
func TestMongoHelper(t *testing.T) {

	db := NewMongo(context.TODO(), "test", User{})
	finded, err := db.FindAll(&bson.D{{"userName", "mehdi"}})
	if err != nil {
		fmt.Println(err)
	}
	users, _ := json.Marshal(finded)
	fmt.Println(string(users))
	tr,err := StartTransaction(dbId, context.Background())
	if err!=nil{
		return 
	}
	transFn:=func(sessionContext mongo.SessionContext) (result interface{}, err error) {
		db := NewMongo(sessionContext, "test", User{Username: "mehdi", Password: "1234"})
		db.Insert()
		db2 := NewMongo(sessionContext, "test2", User{Username: "mate", Password: "1234"})
		res, err := db2.Insert()
		if err == nil {
			return res, errors.New("Abort transaction")
		}
		return res, err
	}
	res,err :=tr.EndTransaction(transFn)
	fmt.Println(res)
}
