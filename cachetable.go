package memredis

import (
	"sync"
	"time"
	"fmt"
)
var Cachetable = &CacheTable{items: make(map[interface{}] *CacheItem)}

type CacheTable struct{
	sync.RWMutex
	items map[interface{}] *CacheItem

	cleanupTimer *time.Timer

	cleanupInterval time.Duration
}

func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.Unlock()
	for k, v := range table.items {
		trans(k, v)
	}
}

func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]
	return ok
}

func (table *CacheTable) Get(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	table.RUnlock()
	if ok {
		r.KeepAlive()
		return r, nil
	}
	return nil, ErrKeyNotFound
}

func (table *CacheTable) Set(key string,  lifeSpan time.Duration, data interface{}) (ok bool){
	table.Lock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheItem(key, lifeSpan, data)
	}else{
		if(item.itemType != ItemType_STRING) {
			return false
		}
		item.lifeSpan = lifeSpan
		item.data = data
		item.KeepAlive()
	}
	table.items[key] = item
	table.Unlock()

	// update expire check time
	if lifeSpan > 0 && (lifeSpan < table.cleanupInterval || table.cleanupInterval == 0) {
		go table.expirationCheck()
	}

	return true
}

func (table *CacheTable) SAdd(key string, lifeSpan time.Duration, data string)(ok bool) {
	table.Lock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheSetItem(key, lifeSpan, data)
	}else{
		if(item.itemType != ItemType_SET){
			return false
		}
		item.lifeSpan = lifeSpan
		dataMap := item.data.(map[string]*Empty)
		dataMap[data] = &Empty{}
		item.KeepAlive()
	}
	table.items[key] = item
	table.Unlock()

	// update expire check time
	if lifeSpan > 0 && (lifeSpan < table.cleanupInterval || table.cleanupInterval == 0) {
		go table.expirationCheck()
	}

	return true

}


func (table *CacheTable) delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()
	return table.deleteInternal(key)

}

// this delete without lock, should be used internal, such as expire check
func (table *CacheTable) deleteInternal(key interface{})(*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	fmt.Println("deletin key:", key, "value: ", r.data)
	delete(table.items, key)
	return r, nil
}

func (table *CacheTable) expirationCheck() {
	table.Lock()
	fmt.Println("expire check start....")
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}

	now := time.Now()
	smallestDuration := 0 * time.Second

	for key, item := range table.items {

		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		// no expire time
		if lifeSpan == 0 {
			continue
		}

		// delete expire key and update the next check time
		if now.Sub(accessedOn) >= lifeSpan {
			table.deleteInternal(key)
		}else {
			if smallestDuration == 0 || lifeSpan - now.Sub(accessedOn) < smallestDuration{
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}
	// update next check time
	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	table.Unlock()
	fmt.Println("expire check end....")
}
