package my_cache

import (
	"log"
	"my_cache/day6_single_flight/my_cache/single_flight"
	"sync"
)

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
	peers     PeerPicker

	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *single_flight.Group
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
		loader:    &single_flight.Group{},
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
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 兜底方案 可以通过用户设置的数据源获取
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

func (g *group) RegisterPeers(peers PeerPicker) {
	//已经存在了
	if g.peers != nil {
		log.Print("RegisterPeerPicker called more than once:", peers)
	}
	g.peers = peers
}

func (g *group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
