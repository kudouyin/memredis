package memredis

import (
	"sync"
	"time"
)

type Empty struct {

}

type ItemType int
const (
	ItemType_STRING ItemType = 0
	ItemType_SET ItemType = 1
	ItemType_LIST ItemType = 2
)

type CacheItem struct {
	sync.RWMutex

	key string
	data interface{}
	lifeSpan time.Duration
	createdOn time.Time
	accessedOn time.Time
	accessCount int64
	itemType ItemType

	aboutToExpire func(key interface{})
}

func NewCacheStringItem(key string, lifeSpan time.Duration, data interface{}) *CacheItem{
	t := time.Now()
	return &CacheItem{
		key: key,
		lifeSpan:lifeSpan,
		createdOn: t,
		accessedOn: t,
		accessCount: 0,
		aboutToExpire: nil,
		data: data,
		itemType: ItemType_STRING,
	}
}

// use map to implement a set(value is empty struct)
func NewCacheSetItem(key string, lifeSpan time.Duration, data string) *CacheItem {
	t := time.Now()
	dataMap := make(map[string]*Empty)
	dataMap[data] = &Empty{}
	return &CacheItem{
		key: key,
		lifeSpan:lifeSpan,
		createdOn: t,
		accessedOn: t,
		accessCount: 0,
		aboutToExpire: nil,
		data: dataMap,
		itemType: ItemType_SET,
	}
}

func (item *CacheItem) accessUpdate() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}

func (item *CacheItem) isExpire() bool {
	// Maybe it is 'createdOn', not accessOn
	if item.lifeSpan != 0 && time.Now().Sub(item.accessedOn) >= item.lifeSpan{
		return true
	}
	return false
}