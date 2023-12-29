package mongoHelper

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Holder *ConnManager

func Init(){
	if Holder==nil{
		Holder = new(ConnManager)
		Holder.holder = make(map[string]*mongo.Client)
	}
}
// type MongoDb struct{
// 	db *mongo.Client
// 	uri string
// }
// type optionFunc func(*MongoInstance)
// func (m *MongoDb)GetDbConnection(dbName string, opts ...*options.DatabaseOptions)*mongo.Client{
// 	if m.db!=nil{
// 		m.connect()
// 	}
// 	return m.db.Database(dbName,opts...).Client()
// }
func connect(uri string)(*mongo.Client,error){
	c,err:= mongo.NewClient(options.Client().ApplyURI(uri).SetMaxPoolSize(500).SetMinPoolSize(2))
	c.Connect(context.Background())
	return c,err
}


func New(Id ,Uri string) *mongo.Client {
	if Holder==nil{
		Init()
	}
	c,err:=connect(Uri)
	if err==nil{
		Holder.Write(Id,c)
	}
	return c
}

