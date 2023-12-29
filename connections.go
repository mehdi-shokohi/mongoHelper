package mongoHelper

import (
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)
type ConnManager struct {
	holder map[string]*mongo.Client
	mutex sync.RWMutex
}


func (d *ConnManager) Read(key string) (*mongo.Client, bool) {
	d.mutex.RLock()

	defer d.mutex.RUnlock()
	val, exists := d.holder[key]
	return val, exists
}

func (d *ConnManager) Write(key string, value *mongo.Client) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.holder[key] = value
}

func (d *ConnManager) Delete(key string)  {
	d.mutex.RLock()

	defer d.mutex.RUnlock()
	delete(d.holder,key)
}