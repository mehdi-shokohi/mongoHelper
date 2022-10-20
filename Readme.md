#### Impl By Struct Model
```go
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

```
#### Impl by Map 

```go


	// Map 
	userMap:=make(map[string]interface{})
	dbMap := mongoHelper.NewMongo(context.Background(),"users", userMap)
	_, err = dbMap.FindOne(&bson.D{{Key: "username", Value: "admin"}, {Key: "password", Value: "1234"}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("user not found ")
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println(userMap["_id"].(primitive.ObjectID).Hex())
	fmt.Println(userMap["username"])
	fmt.Println(userMap["password"])
	fmt.Println(userMap["admin"])
