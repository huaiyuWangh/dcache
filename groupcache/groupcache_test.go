package groupcache

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

var db = map[string]string {
	"Tom": "98",
	"Jack": "95",
	"Sam": "94",
}

func TestGetter(t *testing.T) {
	var f GetterFunc = func(key string) ([]byte, error) {
		return []byte(key), nil
	}
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, []byte("key")) {
		t.Fatal("getter failed")
	}
}

func TestGet(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache := NewGroup("score", 1024, func(key string) ([]byte, error) {
				t.Logf("%s", key)
				return []byte(db[key]), nil
			})
			for k, v := range db {
				cachev, _ := cache.Get(k)
				if cachev.String() != v {
					t.Fatal("get failed")
				}
			}
		}()
	}
	wg.Wait()
}

func TestCurl(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for key, val := range db {
				res, _ := http.Get("http://localhost:8081/api?key="+key)
				body,_ := ioutil.ReadAll(res.Body)
				if string(body) != val {
					t.Logf("get failed")
				}
				res.Body.Close()
			}
		}()
	}
	wg.Wait()
}