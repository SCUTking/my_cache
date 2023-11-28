package single_flight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{} //保存请求结果
	err error       //保存请求的错误
}
type Group struct {
	mu sync.Mutex       //控制并发访问 m
	m  map[string]*call //每个key对应一个call，每个call 又对应一个 waitGroup
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call, 0)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() //等待之前的请求完成，直接获取刚刚请求的返回值
		return c.val, c.err
	}

	//说明之前没有请求
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	//执行获取缓存的请求
	i, err := fn()
	c.val = i
	c.err = err
	c.wg.Done()

	//删除并发临界区要加锁
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
