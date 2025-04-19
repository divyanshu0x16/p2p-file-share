package peer

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"encoding/binary"
	"p2p-file-share/file"
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
		}else if strings.HasPrefix(message, "GET_FILE "){
			fileName := strings.TrimPrefix(message, "GET_FILE ")
			p.sendFile(conn, fileName)
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

func (p *Peer) sendFile(conn net.Conn, fileName string){
	fmt.Printf("Received request for file: %s\n", fileName)

	data, err := file.ReadFile(fileName, p.shareDir)
	if err != nil {
		errorMsg := fmt.Sprintf("ERROR: %v\n", err)
		fmt.Println(errorMsg)
		conn.Write([]byte("ERR!"))
		conn.Write([]byte(errorMsg))
		return
	}

	//Send file size first (4 bytes)
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(data)))
	if _, err := conn.Write(sizeBuf); err != nil {
		fmt.Printf("Failed to send file size: %v\n", err)
		return
	}

	//Send file data
	if _, err := conn.Write(data); err != nil {
		fmt.Printf("Failed to send file: %v\n", err)
		return
	}

	fmt.Printf("Successfully sent file: %s (%d bytes)\n", fileName, len(data))
}