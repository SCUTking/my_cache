package my_cache

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"my_cache/day5_rpc/my_cache/consistent_hash"
	"my_cache/day5_rpc/my_cache/geecachepb"
	"strings"
	"sync"
)

const (
	defaultReplicas = 50
)

// /rpcPool 既具备了提供 RPC 服务的能力，也具备了根据具体的 key
// ，创建 RPC 客户端从远程节点获取缓存值的能力
type rpcPool struct {
	self string
	//实现这个才能编写服务端程序
	geecachepb.UnimplementedGroupCacheServer
	mu         sync.Mutex            //
	peers      *consistent_hash.Map  //保存着节点信息，通过一致性hash选取节点
	rpcGetters map[string]*rpcGetter // 每个节点对应着一个获取对应节点缓存的getter
}

func NewRpcPool(self string) *rpcPool {
	return &rpcPool{rpcGetters: make(map[string]*rpcGetter, 0), self: self}
}

func (r *rpcPool) Get(ctx context.Context, req *geecachepb.Request) (*geecachepb.Response, error) {
	g := req.Group
	key := req.Key
	group := GetGroup(g)
	if group == nil {
		return nil, errors.New("no such group")
	}
	get, err := group.Get(key)
	if err != nil {
		return nil, err
	}
	return &geecachepb.Response{Value: get.ByteView()}, nil
}

type rpcGetter struct {
	baseURL string
}

func (h *rpcGetter) Get(group string, key string) ([]byte, error) {

	//节点的选择
	dial, err := grpc.Dial(h.baseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer dial.Close()
	client := geecachepb.NewGroupCacheClient(dial)
	rep, err := client.Get(context.Background(), &geecachepb.Request{Group: group, Key: key})
	return rep.GetValue(), err
}

func (r *rpcPool) PickPeer(key string) (PeerGetter, bool) {
	peer := r.peers.Get(key)
	host1 := strings.Split(peer, ":")
	host2 := strings.Split(r.self, ":")
	//判断主机是否相同
	if peer != "" && host1[0] != host2[0] {
		return r.rpcGetters[peer], true
	}
	return nil, false
}

// Set 只在程序启动的时候启动一次
func (r *rpcPool) Set(peers ...string) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.peers = consistent_hash.New(defaultReplicas, nil)
	r.peers.Add(peers...)
	r.rpcGetters = make(map[string]*rpcGetter, len(peers))

	for _, peer := range peers {
		r.rpcGetters[peer] = &rpcGetter{baseURL: peer}
	}

}
