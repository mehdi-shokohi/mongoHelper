package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	conf "github.com/mehdi-shokohi/mongoHelper/config"
)

var mongoWriteDb *MongoWriteDB

// MongoWriteDB ...
type MongoWriteDB struct {
	ctx context.Context
	db  chan *mongo.Database
}

// GetConnection Function
func (pool *MongoWriteDB) GetConnection() *mongo.Database {
	fmt.Println("Len Of Connection : ", len(pool.db))
	if pool.db == nil {
		pool.initialPool()
	}
	if len(pool.db) == 0 {
		client, err := mongo.NewClient(options.Client().ApplyURI(conf.GetMongoAddress()).SetMaxPoolSize(500).SetMinPoolSize(2))
		if err != nil {

			panic(err)
		}
		//ctx, cancel := context.WithTimeout(pool.ctx, 10*time.Second)
		//defer cancel()
		err = client.Connect(pool.ctx)
		collection := client.Database(conf.GetMongodbName())

		if err = client.Ping(pool.ctx, readpref.Primary()); err == nil {
			pool.db <- collection
		}
	}

	return <-pool.db
}

// Release function
func (pool *MongoWriteDB) Release(con *mongo.Database) {
	if len(pool.db) > 500 {
		_ = con.Client().Disconnect(pool.ctx)
	} else {
		pool.db <- con
	}
}

//InitialPool
func (pool *MongoWriteDB) initialPool() {
	pool.ctx = context.Background()
	pool.db = make(chan *mongo.Database, 1000)

}
func GetWriteDB() *MongoWriteDB {
	if mongoWriteDb == nil {
		mongoWriteDb = new(MongoWriteDB)
	}
	return mongoWriteDb
}
