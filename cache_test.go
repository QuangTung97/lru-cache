package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUList(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		list := newLruList(5)
		assert.Equal(t, 6, len(list))

		expected := lruEntry{
			next: 0,
			prev: 0,
		}
		assert.Equal(t, expected, list[0])
	})

	t.Run("add", func(t *testing.T) {
		list := newLruList(4)
		list.add(1, 5)

		expected := lruList([]lruEntry{
			{next: 1, prev: 1},
			{next: 0, prev: 0, key: 5},
		})

		assert.Equal(t, expected, list[:2])
		list.add(2, 10)

		expected = lruList([]lruEntry{
			{next: 2, prev: 1},
			{next: 0, prev: 2, key: 5},
			{next: 1, prev: 0, key: 10},
		})
		assert.Equal(t, expected, list[:3])

		list.add(3, 15)
		expected = lruList([]lruEntry{
			{next: 3, prev: 1},
			{next: 0, prev: 2, key: 5},
			{next: 1, prev: 3, key: 10},
			{next: 2, prev: 0, key: 15},
		})
		assert.Equal(t, expected, list[:4])

		list.add(4, 20)
		expected = lruList([]lruEntry{
			{next: 4, prev: 1},
			{next: 0, prev: 2, key: 5},
			{next: 1, prev: 3, key: 10},
			{next: 2, prev: 4, key: 15},
			{next: 3, prev: 0, key: 20},
		})
		assert.Equal(t, expected, list[:5])

		entryID, oldKey := list.updateLast(30)
		assert.Equal(t, 5, oldKey)
		assert.Equal(t, lruEntryID(1), entryID)

		expected = lruList([]lruEntry{
			{next: 1, prev: 2},
			{next: 4, prev: 0, key: 30},
			{next: 0, prev: 3, key: 10},
			{next: 2, prev: 4, key: 15},
			{next: 3, prev: 1, key: 20},
		})
		assert.Equal(t, expected, list[:5])

		list.touch(3)
		expected = lruList([]lruEntry{
			{next: 3, prev: 2},
			{next: 4, prev: 3, key: 30},
			{next: 0, prev: 4, key: 10},
			{next: 1, prev: 0, key: 15},
			{next: 2, prev: 1, key: 20},
		})
		assert.Equal(t, expected, list[:5])
	})
}

func TestSet(t *testing.T) {
	t.Run("init", func(t *testing.T) {
		cache := NewCache(5)
		assert.Equal(t, 0, len(cache.data))

		assert.Equal(t, uint32(5), cache.maxNum)
		assert.Equal(t, uint32(0), cache.num)
	})

	t.Run("set", func(t *testing.T) {
		cache := NewCache(5)

		cache.Set(10, 100)

		expectedData := map[Key]wrappedValue{
			10: {
				entryID: lruEntryID(1),
				value:   100,
			},
		}
		assert.Equal(t, expectedData, cache.data)
		assert.Equal(t, uint32(1), cache.num)
		assert.Equal(t, uint32(5), cache.maxNum)

		expected := lruList([]lruEntry{
			{next: 1, prev: 1},
			{next: 0, prev: 0, key: 10},
		})
		assert.Equal(t, expected, cache.list[:2])

		cache.Set(20, 200)

		expectedData = map[Key]wrappedValue{
			10: {
				entryID: lruEntryID(1),
				value:   100,
			},
			20: {
				entryID: lruEntryID(2),
				value:   200,
			},
		}
		assert.Equal(t, expectedData, cache.data)
		assert.Equal(t, uint32(2), cache.num)
		assert.Equal(t, uint32(5), cache.maxNum)

		expected = lruList([]lruEntry{
			{next: 2, prev: 1},
			{next: 0, prev: 2, key: 10},
			{next: 1, prev: 0, key: 20},
		})
		assert.Equal(t, expected, cache.list[:3])
	})

	t.Run("set-override", func(t *testing.T) {
		cache := NewCache(5)

		cache.Set(10, 100)
		cache.Set(20, 200)
		cache.Set(30, 300)
		cache.Set(40, 400)
		cache.Set(50, 500)

		assert.Equal(t, uint32(5), cache.num)
		assert.Equal(t, uint32(5), cache.maxNum)

		cache.Set(30, 600)
		assert.Equal(t, uint32(5), cache.num)
		assert.Equal(t, uint32(5), cache.maxNum)

		expectedData := map[Key]wrappedValue{
			10: {
				entryID: lruEntryID(1),
				value:   100,
			},
			20: {
				entryID: lruEntryID(2),
				value:   200,
			},
			30: {
				entryID: lruEntryID(3),
				value:   600,
			},
			40: {
				entryID: lruEntryID(4),
				value:   400,
			},
			50: {
				entryID: lruEntryID(5),
				value:   500,
			},
		}
		assert.Equal(t, expectedData, cache.data)

		expected := lruList([]lruEntry{
			{next: 3, prev: 1},
			{next: 0, prev: 2, key: 10},
			{next: 1, prev: 4, key: 20},
			{next: 5, prev: 0, key: 30},
			{next: 2, prev: 5, key: 40},
			{next: 4, prev: 3, key: 50},
		})
		assert.Equal(t, expected, cache.list[:6])
	})

	t.Run("set-evict", func(t *testing.T) {
		cache := NewCache(5)

		cache.Set(10, 100)
		cache.Set(20, 200)
		cache.Set(30, 300)
		cache.Set(40, 400)
		cache.Set(50, 500)

		cache.Set(60, 600)

		assert.Equal(t, uint32(5), cache.num)
		assert.Equal(t, uint32(5), cache.maxNum)

		expectedData := map[Key]wrappedValue{
			20: {
				entryID: lruEntryID(2),
				value:   200,
			},
			30: {
				entryID: lruEntryID(3),
				value:   300,
			},
			40: {
				entryID: lruEntryID(4),
				value:   400,
			},
			50: {
				entryID: lruEntryID(5),
				value:   500,
			},
			60: {
				entryID: lruEntryID(1),
				value:   600,
			},
		}
		assert.Equal(t, expectedData, cache.data)

		expected := lruList([]lruEntry{
			{next: 1, prev: 2},
			{next: 5, prev: 0, key: 60},
			{next: 0, prev: 3, key: 20},
			{next: 2, prev: 4, key: 30},
			{next: 3, prev: 5, key: 40},
			{next: 4, prev: 1, key: 50},
		})
		assert.Equal(t, expected, cache.list[:6])
	})
}

func TestGet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		cache := NewCache(5)

		cache.Set(10, 100)
		cache.Set(20, 200)
		cache.Set(30, 300)
		cache.Set(40, 400)
		cache.Set(50, 500)

		_, ok := cache.Get(60)
		assert.False(t, ok)

		value, ok := cache.Get(40)
		assert.True(t, ok)
		assert.Equal(t, 400, value)

		expected := lruList([]lruEntry{
			{next: 4, prev: 1},
			{next: 0, prev: 2, key: 10},
			{next: 1, prev: 3, key: 20},
			{next: 2, prev: 5, key: 30},
			{next: 5, prev: 0, key: 40},
			{next: 3, prev: 4, key: 50},
		})
		assert.Equal(t, expected, cache.list[:6])
	})
}

func TestGetAll(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		cache := NewCache(5)

		cache.Set(10, 100)
		cache.Set(20, 200)
		cache.Set(30, 300)
		cache.Set(40, 400)
		cache.Set(50, 500)

		expected := map[Key]Value{
			10: 100,
			20: 200,
			30: 300,
			40: 400,
			50: 500,
		}
		assert.Equal(t, expected, cache.GetAll())
	})
}
