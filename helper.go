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

func connect(uri string)(*mongo.Client,error){
	c,err:= mongo.NewClient(options.Client().ApplyURI(uri).SetMaxPoolSize(500).SetMinPoolSize(2))
	if err!=nil{
		return nil,err
	}
	err=c.Connect(context.Background())
	return c,err
}

func GetClientById(key string)(*mongo.Client,bool){
	if Holder==nil{
		Init()
	}
	return Holder.read(key)
}
func New(Id ,Uri string) *mongo.Client {
	if Holder==nil{
		Init()
	}
	c,err:=connect(Uri)
	if err==nil{
		Holder.write(Id,c)
	}
	return c
}

