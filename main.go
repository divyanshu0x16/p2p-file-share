package main

import (
	"fmt"
	"os"
	"p2p-file-share/peer"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <listen-port> <share-directory>")
		fmt.Println("Example: go run main.go 8080 ./shared_files")
		return
	}
	
	port := os.Args[1]
	shareDir := os.Args[2]
	listenAddr := ":" + port

	// Ensure the share directory exists
	if _, err := os.Stat(shareDir); os.IsNotExist(err) {
		fmt.Printf("Share directory '%s' does not exist. Creating it...\n", shareDir)
		err = os.MkdirAll(shareDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
			return
		}
	}
	
	p := peer.NewPeer(listenAddr, shareDir)
	
	go p.StartListening()
	p.StartCommandLine()
}