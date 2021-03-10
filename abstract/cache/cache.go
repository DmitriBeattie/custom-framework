package cache

import (
	"sync"

	"github.com/matryer/resync"
)

type RefreshCacheFunc func() (map[interface{}]interface{}, error)

type RefreshRuleFunc func() bool

type ErrHandler func(err error)

type Cache struct {
	data map[interface{}]interface{}
	m    sync.RWMutex
	f    RefreshCacheFunc
	r    RefreshRuleFunc
	e    ErrHandler
	once resync.Once
}

func NewCache(ref RefreshCacheFunc, ruleFunc RefreshRuleFunc, eHandler ErrHandler) *Cache {
	return &Cache{
		f: ref,
		r: ruleFunc,
		e: eHandler,
	}
}

func (c *Cache) initCache(resetOnce bool) error {
	var err error
	var data map[interface{}]interface{}
	var cacheReseted bool

	c.m.Lock()
	defer c.m.Unlock()

	if resetOnce {
		c.once.Reset()
	}

	c.once.Do(
		func() {
			cacheReseted = true
			data, err = c.f()
		},
	)

	if err == nil && cacheReseted {
		c.data = data
	}

	if err != nil {
		c.e(err)
	}

	return err
}

func (c *Cache) checkNeedResetCache() error {
	c.m.RLock()

	if c.r() {
		c.m.RUnlock()

		return c.initCache(true)
	}

	c.m.RUnlock()

	return nil
}

func (c *Cache) get() (map[interface{}]interface{}, error) {
	if err := c.initCache(false); err != nil {
		return nil, err
	}

	if err := c.checkNeedResetCache(); err != nil {
		return nil, err
	}

	c.m.RLock()

	copiedData := make(map[interface{}]interface{})

	for key, val := range c.data {
		copiedData[key] = val
	}

	c.m.RUnlock()

	return copiedData, nil
}

func (c *Cache) GetAll() (map[interface{}]interface{}, error) {
	return c.get()
}

func (c *Cache) Get(key interface{}) (interface{}, error) {
	data, err := c.get()
	if err != nil {
		return nil, err
	}

	return data[key], nil
}

func (c *Cache) GetOk(key interface{}) (interface{}, bool, error) {
	data, err := c.get()
	if err != nil {
		return nil, false, err
	}

	val, ok := data[key]

	return val, ok, nil
}
