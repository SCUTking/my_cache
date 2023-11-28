package my_cache

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}
type entry struct {
	key   string
	value Value
}
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string2 string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element, 0),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {

	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		//element.Value 是获取双向链表的节点的值
		e := element.Value.(*entry)
		return e.value, true
	} else {
		return nil, false
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//之前已经存在，就是更新操作
		c.ll.MoveToFront(ele) //移动到队列开头
		kv := ele.Value.(*entry)
		//改变长度
		c.nbytes -= int64(kv.value.Len()) - int64(value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nbytes += int64(value.Len()) + int64(len(key))
	}

	//如果出现内存溢出，进行内存淘汰
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest() //循环淘汰直到满足条件
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		e := ele.Value.(*entry)
		delete(c.cache, e.key)
		c.nbytes -= int64(len(e.key)) + int64(e.value.Len())
		//回调函数的   在删除的使用
		if c.OnEvicted != nil {
			c.OnEvicted(e.key, e.value)
		}
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
