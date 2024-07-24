package core

import (
	"ray/config"
	"time"
)

var store map[string]*Obj

func init() {
	store = make(map[string]*Obj)
}

func NewObj(value interface{}, durationMs int64, oType uint8, oEnc uint8) *Obj {

	var expiresAt int64 = -1

	if durationMs > 0 {
		var currTime int64 = time.Now().UnixMilli()
		expiresAt = currTime + durationMs
	}

	return &Obj{
		Value:        value,
		ExpiresAt:    expiresAt,
		TypeEncoding: oType | oEnc,
	}

}

func Put(key string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	store[key] = obj
}

// make eviction strategy configuration driven
// TODO : make it efficient by doing through sampling
func evict() {
	evictFirst()
}

func evictFirst() {
	for key := range store {
		delete(store, key)
		return
	}
}

func Get(key string) *Obj {

	v := store[key]

	// logic for passive delete, where delete when accessed if the key is expired.
	if v != nil && v.ExpiresAt != -1 && v.ExpiresAt < time.Now().UnixMilli() {
		delete(store, key)
		return nil
	}
	return v
}

func Delete(key string) bool {
	if _, ok := store[key]; ok {
		delete(store, key)
		return true
	}
	return false
}
