package server

import (
	"dcache/cachepb"
	"dcache/consistenthash"
	"dcache/groupcache"
	"dcache/remote"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type HTTPPool struct {
	self string
	basePath string
	peers *consistenthash.HashCircle
	httpGetters map[string]*httpGetter
	mu sync.Mutex
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *cachepb.Request, out *cachepb.Response) (err error) {
	 u:= fmt.Sprintf(
		"%v/%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	resp ,err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error %d", resp.StatusCode)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body failed")
	}
	if err = proto.Unmarshal(bytes, out) ;err != nil {
		return
	}
	return
}


const defaultBasePath = "_dcache"
const defaultReplicas = 50

func NewHTTPPool (self string) *HTTPPool {
	return &HTTPPool{
		self:        self,
		basePath:    defaultBasePath,
		httpGetters: make(map[string]*httpGetter),
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if p.self != req.Host {
		http.Error(w, fmt.Sprintf("wrong host %s [%s]", req.Host, p.self), http.StatusBadRequest)
		return
	}
	path := strings.Split(req.URL.Path, "/")
	if len(path) < 4 || path[1] != p.basePath {
		http.Error(w, fmt.Sprintf("wrong path %s", req.URL.Path), http.StatusBadRequest)
		return
	}
	group := groupcache.GetGroup(path[2])
	if group == nil {
		http.Error(w, "wrong cache", http.StatusBadRequest)
		return
	}
	key := path[3]
	res, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	body, err  := proto.Marshal(&cachepb.Response{Value: res.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: fmt.Sprintf("http://%s/%s",peer,p.basePath),
		}
	}
	//p.peers.Print()
	return
}

func (p *HTTPPool) PickPeer(key string) (getter remote.PeerGetter, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.Get(key)
	p.Log("Pick peer %s", peer)
	if peer == "" || peer == p.self {
		return
	}
	getter, ok = p.httpGetters[peer]
	return
}

var _ remote.PeerPicker = (*HTTPPool)(nil)
