package peer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"p2p-file-share/file"
	"strings"
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

func (p *Peer) downloadFile(address string, fileName string){
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Downloading file '%s' from %s\n", fileName, address)

	request := fmt.Sprintf("GET_FILE %s\n", fileName)
	//Send file request
	if _, err := conn.Write([]byte(request)); err != nil {
		fmt.Printf("Failed to send file request: %v\n", err)
		return
	}

	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, sizeBuf); err != nil {
		fmt.Printf("Failed to read file size: %v\n", err)
		return
	}

	if string(sizeBuf) == "ERR!" {
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed to read error message: %v\n", err)
			return
		}

		if strings.HasPrefix(string(response), "ERROR"){
			fmt.Println(string(response))
			return
		}
	}

	fileSize := binary.BigEndian.Uint32(sizeBuf)
	fmt.Printf("Receiving file of size: %d bytes\n", fileSize)

	//Receive the data
	fileData := make([]byte, fileSize)
	bytesRead := 0
	for bytesRead < int(fileSize){
		n, err := conn.Read(fileData[bytesRead:])
		//TODO: Add a timer here to automatically close connection and fail after some time

		if err != nil && err != io.EOF {
			fmt.Printf("Error reading file data: %v\n", err)
			return
		}

		if n == 0 {
			break
		}
		bytesRead += n

		// Print progress
		progress := float64(bytesRead) / float64(fileSize) * 100
		fmt.Printf("\rDownload progress: %.1f%% (%d/%d bytes)", progress, bytesRead, fileSize)
	}
	fmt.Println() //End the progress file

	if err := file.SaveFile(fileData, fileName, p.shareDir); err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		return
	}
	
	fmt.Printf("File downloaded successfully: %s\n", fileName)
	p.scanLocalFiles()
}