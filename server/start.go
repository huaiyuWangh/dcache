package server

import (
	"dcache/groupcache"
	"fmt"
	"log"
	"net"
	"net/http"
)

var db = map[string]string {
	"Tom": "98",
	"Jack": "95",
	"Sam": "94",
}
var localIP string
var groupCache *groupcache.Group

func init() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal("Init IP Error!!")
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP = ipnet.IP.String()
			}
		}
	}
	groupCache = groupcache.NewGroup("score", 1024, func(key string) ([]byte, error) {
		if _, ok := db[key]; ok {
			return []byte(db[key]), nil
		} else {
			return nil, fmt.Errorf("%s not exist", key)
		}
	})
}

func StartAPIServe(port int) {
	handleFunc := func(w http.ResponseWriter, req *http.Request) {
		key := req.URL.Query().Get("key")
		value, err := groupCache.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(value.ByteSlice())
	}
	http.HandleFunc("/api", handleFunc)
	log.Printf("API Serving On %d...",port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}


func StartCacheServe(port int) {
	peers := NewHTTPPool(fmt.Sprintf("%s:%d", localIP, port))
	groupCache.RegisterPeers(peers)
	go Register(port)
	go Discover(peers)
	log.Printf("Cache Serving On %d...",port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), peers))
}