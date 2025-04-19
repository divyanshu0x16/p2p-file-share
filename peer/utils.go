package peer

import (
	"p2p-file-share/file"
	"fmt"
)

func (p *Peer) scanLocalFiles() {
	fileList, err := file.ScanDirectory(p.shareDir, p.ListenAddr)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
	}
	
	p.localFiles = fileList
	fmt.Printf("Scanned %d files in %s\n", len(fileList.Files), p.shareDir)
}

func (p *Peer) addPeer(address string){
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.peers[address] = true
	fmt.Println("Current peers:", p.peers)
}