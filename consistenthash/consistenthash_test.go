package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashCircle(t *testing.T) {
	hosts := []string{"1","2","3","4","5"}
	hc := New(5, func(bytes []byte) uint32 {
		i, _ :=  strconv.Atoi(string(bytes))
		return uint32(i)
	})
	hc.Add(hosts...)

	if host := hc.Get("10"); host != "1" {
		t.Fatal("get 10 failed")
	}
	if host := hc.Get("15"); host != "2" {
		t.Fatal("get 11 failed")
	}
	if host := hc.Get("55"); host != "1" {
		t.Fatal("get 55 failed")
	}
}
