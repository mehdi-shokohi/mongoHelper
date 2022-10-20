#### Impl By Struct Model
```go
	config.SetConfig(config.Config{
		MongoAddress: "mongodb://localhost:27018/?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false",
		MongoDbName: "goex" ,
		}) // run once . In Main Func


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
##### Extend mongo Helper 
```go
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

```
### Test Of Extended MongoHelper
```go
	// Extended Helper
	db2 := MyNewMongo(context.Background(), "users", &Model{})
	userModel2, err2 := db2.MyFindOne(&bson.D{{Key: "username", Value: "naser"}, {Key: "password", Value: "1234"}})
	if err2 != nil {
		if err2 == mongo.ErrNoDocuments {
			fmt.Println("user not found ")
			return 
		} else {
			fmt.Println(err.Error())
			return 
		}
	}

	fmt.Println(userModel2.ID.Hex())
	fmt.Println(userModel2.Username)
	fmt.Println(userModel2.Password)
	fmt.Println(userModel2.Admin)
	

