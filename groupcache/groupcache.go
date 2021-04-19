package groupcache

import (
	"dcache/byteview"
	"dcache/cachepb"
	"dcache/local"
	"dcache/remote"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache *local.Cache
	picker    remote.PeerPicker
	loader    *G
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, maxBytes int64, getter GetterFunc) *Group {
	if getter == nil {
		panic("getter nil")
	}
	mu.Lock()
	defer mu.Unlock()
	if g, ok := groups[name]; ok {
		return g
	}
	g := &Group{
		name: name,
		getter: getter,
		mainCache: &local.Cache{MaxBytes: maxBytes},
		loader: new(G),
	}
	groups[name] = g
	return g
}

func (g *Group) RegisterPeers(picker remote.PeerPicker) {
	g.picker = picker
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (bv byteview.ByteView, err error) {
	bv, ok := g.mainCache.Get(key)
	if ok {
		return
	}
	val, err := g.loader.Do(key, func() (interface{}, error) {
		return g.load(key)
	})
	if err != nil {
		return
	}
	return val.(byteview.ByteView), nil
}

func (g *Group) load(key string) (bv byteview.ByteView, err error) {
	if g.picker != nil {
		if peer, ok := g.picker.PickPeer(key); ok {
			req := &cachepb.Request{
				Group: g.name,
				Key: key,
			}
			res := &cachepb.Response{}
			if err := peer.Get(req, res); err != nil {
				return bv, err
			}
			return byteview.ByteView{B: res.GetValue()}, nil
		}
	}
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (bv byteview.ByteView, err error) {
	bytes , err := g.getter.Get(key)
	if err != nil {
		return
	}
	bv = byteview.ByteView{B: bytes}
	g.mainCache.Set(key, bv)
	return
}


