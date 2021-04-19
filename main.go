package main

import (
	"dcache/server"
	"flag"
)

func main() {
	port := flag.Int("port", 10086, "")
	apiport := flag.Int("apiport", 0, "")
	flag.Parse()
	if *apiport > 0 {
		go server.StartAPIServe(*apiport)
	}
	server.StartCacheServe(*port)
}
