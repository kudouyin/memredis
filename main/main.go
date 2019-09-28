package main

import (
	"flag"
	"fmt"
	"github.com/kudouyin/memredis"
)

func main() {
	port := flag.Int("port", 3001, "main port")
	gossipPort := flag.Int("gossipport", 8008, "gossip port")
	seedNodeAddr := flag.String("seednodeaddr", "127.0.0.1:8008", "gossip seed node addr")
	flag.Parse()
	fmt.Println(*seedNodeAddr)
	memredis.Run(port, gossipPort, seedNodeAddr)
}
