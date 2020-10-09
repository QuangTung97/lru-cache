package lru

import "github.com/cheekybits/genny/generic"

// Key type for key
type Key generic.Type

// Value type for value
type Value generic.Type

type lruEntry struct {
	next uint32
	prev uint32
	key  Key
}

type lruList []lruEntry

type lruEntryID uint32

type wrappedValue struct {
	entryID lruEntryID
	value   Value
}

// Cache type for cache
type Cache struct {
	list   lruList
	maxNum uint32
	num    uint32
	data   map[Key]wrappedValue
}

func newLruList(maxNum uint32) lruList {
	list := make([]lruEntry, maxNum+1)

	list[0].next = 0
	list[0].prev = 0

	return list
}

func (l lruList) updateLast(key Key) (lruEntryID, Key) {
	lastID := l[0].prev
	prevLastID := l[lastID].prev

	oldKey := l[lastID].key

	l[prevLastID].next = 0
	l[0].prev = prevLastID

	firstID := l[0].next
	l[0].next = lastID

	l[firstID].prev = lastID

	l[lastID].next = firstID
	l[lastID].prev = 0

	l[lastID].key = key

	return lruEntryID(lastID), oldKey
}

func (l lruList) touch(id lruEntryID) {
	prevID := l[id].prev
	nextID := l[id].next

	l[prevID].next = nextID
	l[nextID].prev = prevID

	firstID := l[0].next
	l[0].next = uint32(id)
	l[firstID].prev = uint32(id)

	l[id].next = firstID
	l[id].prev = 0
}

func (l lruList) add(id lruEntryID, key Key) {
	l[id].prev = 0
	l[id].key = key

	firstID := l[0].next
	l[id].next = firstID
	l[0].next = uint32(id)

	l[firstID].prev = uint32(id)
}

// NewCache returns a new cache
func NewCache(maxNum uint32) *Cache {
	return &Cache{
		list:   newLruList(maxNum),
		maxNum: maxNum,
		num:    0,
		data:   make(map[Key]wrappedValue),
	}
}

// Set add or replace key with value, change position in LRU list
func (c *Cache) Set(key Key, value Value) {
	old, existed := c.data[key]
	if existed {
		c.list.touch(old.entryID)
		c.data[key] = wrappedValue{
			entryID: old.entryID,
			value:   value,
		}
	} else {
		if c.num == c.maxNum {
			entryID, oldKey := c.list.updateLast(key)

			delete(c.data, oldKey)

			c.data[key] = wrappedValue{
				entryID: entryID,
				value:   value,
			}
		} else {
			c.num++
			entryID := lruEntryID(c.num)

			c.list.add(entryID, key)
			c.data[key] = wrappedValue{
				entryID: entryID,
				value:   value,
			}
		}
	}
}

// Get returns a value if existed, change position in LRU list
func (c *Cache) Get(key Key) (value Value, ok bool) {
	v, existed := c.data[key]
	if !existed {
		ok = false
		return
	}

	c.list.touch(v.entryID)

	value = v.value
	ok = true
	return
}

// GetAll returns all key value pairs, not change positions in LRU list
func (c *Cache) GetAll() map[Key]Value {
	result := make(map[Key]Value)
	for k, v := range c.data {
		result[k] = v.value
	}
	return result
}
