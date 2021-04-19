package groupcache

import "sync"

type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

type G struct {
	mu sync.Mutex
	m map[string]*call
}

func (g *G) Do(key string, fn func()(interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	// 延迟初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// call已发起则等待完成
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	// 创建call
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	// 发起请求
	c.val, c.err = fn()
	c.wg.Done()

	// 移除call
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
