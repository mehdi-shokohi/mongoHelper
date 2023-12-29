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
var db *mongo.Client
func NewMongo[T any](ctx context.Context, colName string, model T) MongoContainer[T] {
	if db==nil{

		db=New("mongo_rnj_rw#1","mongodb://localhost:27018/rainjoy?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false")
	}
	m := MongoContainer[T]{}
	m.Model = model
	m.Ctx = ctx
	m.DatabaseName = "rainjoy"
	m.Connection ,_= Holder.Read("mongo_rnj_rw#1")
	m.CollectionName = colName
	return m
}
func TestMongoHelper(t *testing.T) {

db:=NewMongo(context.TODO(),"test",User{})
finded,err:=db.FindAll(&bson.D{{"userName","mehdi"}})
if err!=nil{
	fmt.Println(err)
}
users,_:=json.Marshal(finded)
fmt.Println(string(users))
tr:=StartTransaction("mongo_rnj_rw#1",context.Background())
res,err :=tr.EndTransaction(func(sessionContext mongo.SessionContext) (result interface{}, err error) {
	db:=NewMongo(sessionContext,"test",User{Username: "mehdi",Password: "1234"})
	db.Insert()
	db2:=NewMongo(sessionContext,"test2",User{Username: "mate",Password: "1234"})
	res,err:= db2.Insert()
	if err==nil{
		return res,errors.New("Abort transaction")
	}
	return res,err
})


fmt.Println(res)
}
