package my_cache

import "sync"

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

type group struct {
	name      string //缓存的命名空间
	getter    Getter //缓存未命中时获取源数据的回调(callback)，每个命名空间一个回调
	mainCache cache  //
}

var (
	mu     sync.RWMutex                 //读写锁 用于操作groups时的并发安全
	groups = make(map[string]*group, 0) //用于根据名字进行获取内存的映射
)

func NewGroup(name string, cacheBytes int64, getter Getter) *group {
	if getter == nil {
		panic("nil getter")
	}

	//加上写锁
	mu.Lock()
	defer mu.Unlock()

	newGroup := &group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = newGroup
	return newGroup

}

func GetGroup(name string) *group {
	//加的是读锁
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, nil
	}
	//如果缓存存在，直接返回
	if value, ok := g.mainCache.get(key); ok {
		return value, nil
	}
	//如果缓存不存在
	return g.load(key)
}

func (g *group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, err
}

// 更新获取的信息到缓存中
func (g *group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
