package consistenthash

import (
	"fmt"
	"hash/crc32"
	"log"
	"sort"
)

type HashFunc func([]byte) uint32
var defaultFunc HashFunc = crc32.ChecksumIEEE

type HashCircle struct {
	hash HashFunc
	nodes []int //存放hash环
	hmap map[int]string //hash => host
	replicas int
}

func New(replicas int, hash HashFunc) *HashCircle {
	if hash == nil {
		hash = defaultFunc
	}
	return &HashCircle{
		hash: hash,
		nodes: make([]int, 0),
		hmap: make(map[int]string),
		replicas: replicas,
	}
}

func (h *HashCircle) Add(hosts ...string) {
	for _, host := range hosts {
		for i:=0; i<h.replicas; i++ {
			vhost := fmt.Sprintf("%s%d",host,i)
			hash := int(h.hash([]byte(vhost)))
			h.nodes = append(h.nodes, hash)
			h.hmap[hash] = host
		}
	}
	sort.Ints(h.nodes)
}

func (h *HashCircle) Get(key string) string {
	hash := int(h.hash([]byte(key)))
	index := h.search(hash)
	log.Println("index: ", index)
	return h.hmap[h.nodes[index]]
}

func (h *HashCircle) search(hash int) (index int) {
	index = sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i] >= hash
	})
	return index % len(h.nodes)
}

func (h *HashCircle) Print() {
	hosts := make([]string, 0)
	for _, hash := range h.nodes {
		hosts = append(hosts, h.hmap[hash])
	}
	log.Printf("[hash] %s\n", hosts)
}