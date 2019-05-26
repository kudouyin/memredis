package memredis

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex

	key interface{}
	data interface{}
	lifeSpan time.Duration
	createdOn time.Time
	accessedOn time.Time
	accessCount int64

	aboutToExpire func(key interface{})
}

func NewCacheItem(key interface{}, lifeSpan time.Duration, data interface{}) *CacheItem{
	t := time.Now()
	return &CacheItem{
		key: key,
		lifeSpan:lifeSpan,
		createdOn: t,
		accessedOn: t,
		accessCount: 0,
		aboutToExpire: nil,
		data: data,
	}
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}