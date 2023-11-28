package consistent_hash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map 核心结构
type Map struct {
	hash     Hash           //使用的hash函数，依赖注入，有用户自己传进来
	replicas int            //重复数，如果虚拟节点的数量
	keys     []int          //hash环，用切片模拟环，切片的每一个元素都是一个节点
	hashMap  map[int]string //利用map记录虚拟节点与实际节点的对应关系
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string, 0),
	}
	if m.hash == nil {
		//默认是这个hash函数
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	//key就是实际的节点名称
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//虚拟节点的名字
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	//排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	//
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//环形结构，如果idx在keys里面都不满足，就回到第一个
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}
