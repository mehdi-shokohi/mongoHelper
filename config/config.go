package config
import "errors"

var conf Config

func GetMongodbName()  (mongodbName string) {
	mongodbName = conf.MongoDbName
	if mongodbName == "" {
		panic(errors.New("MONGODB_NAME not found in .env file"))
	}
	return 
}


func GetMongoAddress() (mongoAddress string) {
	mongoAddress = conf.MongoAddress
	if mongoAddress == "" {
		panic(errors.New("MONGO_ADDRESS not found in .env file"))
	}
	return 
}
// func get_env_value(key string) string {
// 	return os.Getenv(key)
// }


func SetConfig(config Config){
	conf = config
}