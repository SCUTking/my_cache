package my_cache

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 这个相当于rpc的客户端
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
