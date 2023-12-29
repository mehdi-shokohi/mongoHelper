#### Impl By Struct Model
if you are need having multi database connection or isolation read and write db connection this repo will be efficient.

```go
	const uri = "mongodb://localhost:27018/?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false"
	const dbId = "mongo_rw_partition_1"
	const database = "contents"

	userModel := new(Model)

	db:=NewMongoById(ctx,dbId,uri,database,"users", userModel)
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
if your env requirements is a db resource, using function same below  

```go

func NewMongo[T any](ctx context.Context,collection string , model T)MongoContainer[T]{
	db:=NewMongoById(ctx,dbId,uri,database,collection,model)
	return db
}

```
#### Impl by Map 

```go


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

```

	

