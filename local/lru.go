package local

import "container/list"

type lru struct {
	maxBytes int64
	nbytes int64
	l *list.List
	m map[string]*list.Element
	onPurged func(string, Value)
}

type entry struct {
	key   string
	value Value
}
func (e *entry) Len() int64 {
	return int64(len(e.key) + e.value.Len())
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onPurged func(string, Value)) *lru {
	return &lru{
		maxBytes: maxBytes,
		l: list.New(),
		m: make(map[string]*list.Element),
		onPurged: onPurged,
	}
}

func (c *lru) purge() {
	ele := c.l.Back()
	if ele == nil {
		return
	}
	c.l.Remove(ele)
	ent := ele.Value.(*entry)
	delete(c.m, ent.key)
	c.nbytes -= ent.Len()
	if c.onPurged != nil {
		c.onPurged(ent.key, ent.value)
	}
}

func (c *lru) Get(key string) (value Value, ok bool) {
	if ele, ok := c.m[key]; ok {
		c.l.MoveToFront(ele)
		ent := ele.Value.(*entry)
		return ent.value, true
	}
	return
}

func (c *lru) Set(key string, value Value) {
	if ele, ok := c.m[key]; ok {
		c.l.MoveToFront(ele)
		ent := ele.Value.(*entry)
		c.nbytes += int64(value.Len() - ent.value.Len())
	} else {
		ent := &entry{key: key, value: value}
		ele = c.l.PushFront(ent)
		c.m[key] = ele
		c.nbytes += ent.Len()
	}

	for c.maxBytes < c.nbytes {
		c.purge()
	}
}

func (c *lru) Len() int {
	return c.l.Len()
}
