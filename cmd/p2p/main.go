package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"example.com/p2p/pkg/message"
	"example.com/p2p/pkg/peer"
)

type addrList []string

func (a *addrList) String() string     { return strings.Join(*a, ",") }
func (a *addrList) Set(s string) error { *a = append(*a, s); return nil }

func main() {
	var peers addrList
	addr := flag.String("addr", "localhost:0", "listen address")
	flag.Var(&peers, "peer", "peer address to connect to (may be repeated)")
	flag.Parse()

	p := peer.New(*addr)
	ln, err := net.Listen("tcp", p.Addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	fmt.Printf("Peer ID %s listening on %s\n", p.ID, ln.Addr().String())

	go func() {
		if err := p.Serve(ln); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	for _, pa := range peers {
		if _, err := p.Connect(pa); err != nil {
			log.Printf("connect %s: %v", pa, err)
		}
	}

	go func() {
		for m := range p.Messages {
			fmt.Printf("[%s] %s\n", m.SenderID, m.Payload)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	seq := 1
	for scanner.Scan() {
		text := scanner.Text()
		msg := &message.Message{SenderID: p.ID, SequenceNo: seq, Payload: text}
		seq++
		if err := p.Broadcast(msg); err != nil {
			log.Printf("broadcast: %v", err)
		}
	}
}
