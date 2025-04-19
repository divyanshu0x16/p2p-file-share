package peer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"p2p-file-share/file"
)

type Peer struct {
	ListenAddr string
	peers map[string]bool
	mutex sync.Mutex
	shareDir string
	localFiles *file.FileList
	remoteFiles map[string]*file.FileList //Map peerId to file list
}

func NewPeer(listenAddr string, shareDir string) *Peer {
	p := &Peer{
		ListenAddr: listenAddr,
		peers: make(map[string]bool),
		shareDir: shareDir,
		remoteFiles: make(map[string]*file.FileList),
	}

	if shareDir != ""{
		p.scanLocalFiles()
	}

	return p
}

// ScanLocalFiles scans the share directory and updates the local file list
func (p *Peer) scanLocalFiles() {
	fileList, err := file.ScanDirectory(p.shareDir, p.ListenAddr)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
		return
	}
	
	p.localFiles = fileList
	fmt.Printf("Scanned %d files in %s\n", len(fileList.Files), p.shareDir)
}

func (p *Peer) StartListening(){
	listener, err := net.Listen("tcp", p.ListenAddr)

	if err != nil {
		fmt.Println("Error starting peer listener:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Listening for connections on", p.ListenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go p.handleConnection(conn)
	}
}

func (p *Peer) StartCommandLine(){
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("P2P File Sharing - Commands:")
	fmt.Println("  connect <address> - Connect to a peer")
	fmt.Println("  list-peers - List known peers")
	fmt.Println("  list-files - List local files")
	fmt.Println("  get-files <address> - Get file list from a peer")
	fmt.Println("  refresh - Rescan the local share directory")
	fmt.Println("  exit - Exit the application")

	for scanner.Scan() {
		command := scanner.Text()
		parts := strings.Fields(command)

		if len(parts) == 0 {
			continue
		}

		switch parts[0]{
		case  "connect":
			if len(parts) < 2 {
				fmt.Println("Usage: connect <address>")
				continue
			}
			go p.connectToPeer(parts[1])

		case "list-peers":
			p.mutex.Lock()
			fmt.Println("Known peers:")
			for peer := range p.peers {
				fmt.Println(" -", peer)
			}
			p.mutex.Unlock()

		case "list-files":
			if p.localFiles == nil {
				fmt.Println("No local files scanned")
			} else {
				file.PrintFileList(p.localFiles)
			}

		case "get-files":
			if len(parts) < 2 {
				fmt.Println("Usage: get-files <address>")
				continue
			}
			go p.requestFileList(parts[1])

		case "refresh":
			p.scanLocalFiles()

		case "exit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Unknown command. Available commands: connect, list, exit")
		}
	}
}

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

// sendFileList sends the local file list to a peer
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

func (p *Peer) addPeer(address string){
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.peers[address] = true
	fmt.Println("Current peers:", p.peers)
}

