package local

import "testing"

type String string
func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(100), nil)
	lru.Set("key1", String("value1"))
	if value, ok := lru.Get("key1"); !ok || value.(String) != "value1" {
		t.Fatalf("cache hit key1=value1 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestEvict(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := String("value1"), String("value1"), String("v3")
	lru := New(int64(len(k1 + k2) + v1.Len() + v2.Len()), nil)
	lru.Set(k1, v1)
	lru.Set(k2, v2)
	lru.Set(k3, v3)
	if _, ok := lru.Get(k1); ok || lru.Len() != 2 {
		t.Fatal("key1 evict failed")
	}

	lru.Get(k2)
	lru.Set(k1, v1)
	v, _ := lru.Get(k2)
	if _, k3ok := lru.Get(k3); k3ok || v.(String) != v2{
		t.Fatal("lru failed")
	}
}
