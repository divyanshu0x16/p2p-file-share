package main

import (
	"fmt"
	"os"
	"p2p-file-share/peer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <listen-port>")
		fmt.Println("Example: go run main.go 8080")
		return
	}
	
	port := os.Args[1]
	listenAddr := ":" + port
	
	p := peer.NewPeer(listenAddr)
	
	go p.StartListening()
	p.StartCommandLine()
}