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