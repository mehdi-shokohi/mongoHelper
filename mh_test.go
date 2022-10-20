package main

import "testing"
import "github.com/mehdi-shokohi/mongoHelper/config"
import "go.mongodb.org/mongo-driver/bson/primitive"
import "go.mongodb.org/mongo-driver/bson"
import "go.mongodb.org/mongo-driver/mongo"

import "github.com/mehdi-shokohi/mongoHelper/mongoHelper"
import "context"
import "fmt"

// User Database Model 
type Model struct {
	ID          *primitive.ObjectID `json:"-"   bson:"_id,omitempty"`
	FirstName   string              `json:"first_name"   bson:"first_name"`
	LastName    string              `json:"last_name"   bson:"last_name"`
	Username    string              `json:"username"   bson:"username"`
	Password    string              `json:"password"   bson:"password"`
	Admin       bool                `json:"admin"   bson:"admin"`
}




func TestMongoHelper(t *testing.T) {
	config.SetConfig(config.Config{
		MongoAddress: "mongodb://localhost:27018/?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false",
		MongoDbName: "goex" ,
		}) // once run . In Main Func

	userModel := new(Model)
	db := mongoHelper.NewMongo(context.Background(),"users", userModel)
	_, err := db.FindOne(&bson.D{{Key: "username", Value: "admin"}, {Key: "password", Value: "1234"}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("user not found ")
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println(userModel.ID.Hex())
	fmt.Println(userModel.Username)
	fmt.Println(userModel.Password)
	fmt.Println(userModel.Admin)


	// Map 
	dbMap := mongoHelper.NewMongo(context.Background(),"users", map[string]interface{}{})
	userMap, err := dbMap.FindAll(&bson.D{})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("user not found ")
		} else {
			fmt.Println(err.Error())
		}
	}
	fmt.Println(len(userMap))
	for _,v:=range userMap {
		fmt.Println((*v)["_id"].(primitive.ObjectID).Hex())
		fmt.Println((*v)["username"])
		fmt.Println((*v)["password"])
		fmt.Println((*v)["admin"])
	}


}
