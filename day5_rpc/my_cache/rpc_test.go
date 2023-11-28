package my_cache

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"my_cache/day5_rpc/my_cache/geecachepb"
	"net"
	"net/http"
	"testing"
)

/*
   $ curl "http://localhost:9999/api?key=Tom"
   630

   $ curl "http://localhost:9999/api?key=kkk"
   kkk not exist
*/

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *group {
	return NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addrs []string, gee *group) {

	for _, s := range addrs {
		listener, err := net.Listen("tcp", s[9:])
		if err != nil {
			panic(err)
		}

		//每个节点都应该
		peers := NewRpcPool(s)

		//一个系统只有一份
		peers.Set(addrs...)
		//gee为最后一份的
		gee.RegisterPeers(peers)

		// 创建grpc server
		server := grpc.NewServer()

		geecachepb.RegisterGroupCacheServer(server, peers)
		//不能阻塞住
		log.Println("cache server is running at port", s)
		go server.Serve(listener)
	}

}

func startAPIServer(apiAddr string) {
	http.Handle("/_geecache", NewHttpApi(apiAddr[16:]))
	log.Println("fontend server is running at", apiAddr)
	//localhost：9999
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func TestRpc(t *testing.T) {

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gee := createGroup()

	startCacheServer(addrs, gee)

	startAPIServer(apiAddr)
}
