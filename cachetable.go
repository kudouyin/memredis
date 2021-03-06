package memredis

import (
	"encoding/json"
	"sync"
	"time"
	"fmt"
)

type CacheTable struct{
	sync.RWMutex
	items map[interface{}] *CacheItem

	cleanupTimer *time.Timer

	cleanupInterval time.Duration

	cleanupNum int
}

func NewCacheTable() *CacheTable{
	cacheTable := &CacheTable{
		items: make(map[interface{}] *CacheItem),
		cleanupInterval: 100000 * time.Millisecond,
		cleanupNum: 100,
	}
	cacheTable.cleanupTimer = time.AfterFunc(cacheTable.cleanupInterval, func(){
		go cacheTable.expirationCheckAll()
	})
	return cacheTable
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

func (table *CacheTable) GetItemInternal(key interface{}, args ...interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if ok {
		// 惰性删除
		if r.isExpire() {
			table.deleteInternal(key)
			return nil, ErrKeyNotFound
		}
		r.accessUpdate()
		return r, nil
	}
	return nil, ErrKeyNotFound
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

//func (table *CacheTable) expirationCheck() {
//	table.Lock()
//	fmt.Println("expire check start....")
//	if table.cleanupTimer != nil {
//		table.cleanupTimer.Stop()
//	}
//
//	now := time.Now()
//	smallestDuration := 0 * time.Second
//
//	for key, item := range table.items {
//
//		item.RLock()
//		lifeSpan := item.lifeSpan
//		accessedOn := item.accessedOn
//		item.RUnlock()
//
//		// no expire time
//		if lifeSpan == 0 {
//			continue
//		}
//
//		// delete expire key and update the next check time
//		if now.Sub(accessedOn) >= lifeSpan {
//			table.deleteInternal(key)
//		}else {
//			if smallestDuration == 0 || lifeSpan - now.Sub(accessedOn) < smallestDuration{
//				smallestDuration = lifeSpan - now.Sub(accessedOn)
//			}
//		}
//	}
//	// update next check time
//	table.cleanupInterval = smallestDuration
//	if smallestDuration > 0 {
//		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
//			go table.expirationCheck()
//		})
//	}
//	table.Unlock()
//	fmt.Println("expire check end....")
//}

func (table *CacheTable) expirationCheckAll() {
	table.Lock()
	defer table.Unlock()

	fmt.Println("expiration check start")
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	cnt := 0
	for key, item := range table.items {
		if item.lifeSpan != 0 {
			cnt += 1
			// random check cleanupNum's key
			if cnt <= table.cleanupNum {
				if item.isExpire() {
					table.deleteInternal(key)
				}
			}
		}
	}
	table.cleanupTimer = time.AfterFunc(table.cleanupInterval, func(){
		go table.expirationCheckAll()
	})
	fmt.Println("expiration check end")
}


// string type command
func (table *CacheTable) Set(key string,  lifeSpan time.Duration, data interface{}) (ok bool, info string){
	table.Lock()
	defer table.Unlock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheStringItem(key, lifeSpan, data)
	}else{
		if(item.itemType != ItemType_STRING) {
			return false, "类型不匹配"
		}
		item.lifeSpan = lifeSpan
		item.data = data
		item.accessUpdate()
	}
	table.items[key] = item

	return true, ""
}

func (table *CacheTable) SETNX(key string, value interface{}) (bool, string){
	table.Lock()
	defer table.Unlock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheStringItem(key, 0, value)
		table.items[key] = item
		return true, ""
	}
	return true, ""
}

func (table *CacheTable) Get(key string) (bool, string) {
	table.RLock()
	item, err := table.GetItemInternal(key)
	table.RUnlock()
	if err != nil {
		fmt.Println(err)
		return false, err.Error()
	}
	fmt.Println("get result before: ", item.data)
	data, ok := item.data.(string)
	if !ok {
		errorInfo := key + "is not a string value type"
		fmt.Println(errorInfo)
		return false, errorInfo
	}
	return true, data
}

// set type command
func (table *CacheTable) SAdd(key string, lifeSpan time.Duration, data string)(ok bool, info string) {
	table.Lock()
	defer table.Unlock()
	item, ok := table.items[key]
	if !ok {
		item = NewCacheSetItem(key, lifeSpan, data)
	}else{
		if(item.itemType != ItemType_SET){
			return false, "类型不匹配"
		}
		item.lifeSpan = lifeSpan
		dataMap := item.data.(map[string]*Empty)
		dataMap[data] = &Empty{}
		item.accessUpdate()
	}
	table.items[key] = item

	return true, ""

}

func (table *CacheTable) SMEMBERS(key string) (bool, string) {
	table.RLock()
	item, err := table.GetItemInternal(key)
	table.RUnlock()
	if err != nil {
		return false, err.Error()
	}
	data, ok := item.data.(map[string]*Empty)
	if !ok {
		errorInfo := key + " is not a set value type"
		fmt.Println(errorInfo)
		return false, errorInfo
	}
	keySet := make([]string, 0, len(data))
	for k := range data {
		keySet = append(keySet, k)
	}
	mjson, err := json.Marshal(keySet)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("get result: ", string(mjson))
	return true, string(mjson)
}

func (table *CacheTable) SPOP(key string)(bool, string) {
	table.Lock()
	defer table.Unlock()
	item, err := table.GetItemInternal(key)
	if err != nil {
		return false, err.Error()
	}
	data, ok := item.data.(map[string]*Empty)
	if !ok {
		errorInfo := key + " is not a set value type"
		fmt.Println(errorInfo)
		return false, errorInfo
	}
	for k := range data {
		delete(data, k)
		return true, k
	}
	return false, ""
}
