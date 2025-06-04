package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"example.com/p2p/pkg/peer"
)

func main() {
	addr := flag.String("addr", "localhost:0", "listen address")
	flag.Parse()

	p := peer.New(*addr)
	ln, err := net.Listen("tcp", p.Addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	fmt.Printf("Peer ID %s listening on %s\n", p.ID, ln.Addr().String())
	select {}
}
