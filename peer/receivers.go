package peer

import (
	"fmt"
	"net"
	"bufio"
	"strings"
)

func (p *Peer) handleConnection(conn net.Conn){
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("New connection from:", remoteAddr)

	p.addPeer(remoteAddr)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Printf("Message from %s: %s\n", remoteAddr, message)

		if strings.HasPrefix(message, "GET_FILES"){
			p.sendFileList(conn)
		}else{
			conn.Write([]byte("Received your message: " + message + "\n"))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

func (p *Peer) sendFileList(conn net.Conn) {
	if p.localFiles == nil {
		conn.Write([]byte("No files available\n"))
		return
	}
	
	jsonData, err := p.localFiles.ToJson()
	if err != nil {
		fmt.Println("Error converting file list to JSON:", err)
		conn.Write([]byte("Error preparing file list\n"))
		return
	}
	
	// Send the JSON data with a newline
	conn.Write(append(jsonData, '\n'))
}