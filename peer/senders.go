package peer

import (
	"net"
	"fmt"
	"strings"
	"bufio"
	"p2p-file-share/file"
)

func (p *Peer) connectToPeer(address string){
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}

	p.addPeer(address)

	fmt.Println("Connected to peer:", address)
	conn.Write([]byte("Hello from " + p.ListenAddr + "\n"))

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
	} else {
		fmt.Println("Response:", strings.TrimSpace(response))
	}

	conn.Close()
}

func (p *Peer) requestFileList(address string){
	conn, err := net.Dial("tcp", address)

	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}

	defer conn.Close()
	p.addPeer(address)

	fmt.Println("Requesting file list from peer:", address)
	//Send file list request
	conn.Write([]byte("GET_FILES\n"))

	//Read response using a buffer
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading file list:", err)
		return
	}

	fileList, err := file.FromJSON([]byte(response))
	if err != nil {
		fmt.Println("Error parsing file list:", err)
		fmt.Println("Raw response:", response)
		return
	}

	//Store the file list
	p.mutex.Lock()
	p.remoteFiles[address] = fileList
	p.mutex.Unlock()

	//Print the file list
	file.PrintFileList(fileList)
}