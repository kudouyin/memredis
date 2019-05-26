package memredis

import (
	"sync"
	"time"
)
var Cachetable = &CacheTable{items: make(map[interface{}] *CacheItem)}

type CacheTable struct{
	sync.RWMutex
	items map[interface{}] *CacheItem
}

func (table *CacheTable) Foreach(trans func(key interface{}, item *CacheItem)) {
	table.RLock()
	defer table.Unlock()
	for k, v := range table.items {
		trans(k, v)
	}
}


func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, data interface{}){
	item := NewCacheItem(key, lifeSpan, data)
	table.Lock()
	table.items[key] = item
	table.Unlock()
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

func (table *CacheTable) Set(key interface{},  lifeSpan time.Duration, data interface{}) (ok bool){
	table.Lock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheItem(key, lifeSpan, data)
	}else{
		item.data = data
	}
	table.items[key] = item
	table.Unlock()
	return true
}
