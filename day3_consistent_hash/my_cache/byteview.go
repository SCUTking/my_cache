package my_cache

// ByteView 抽象一个只读数据结构 ByteView 用来表示缓存
type ByteView struct {
	b []byte
}

// Len 实现Value的接口
func (b ByteView) Len() int {
	return len(b.b)
}

func (v ByteView) ByteView() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
