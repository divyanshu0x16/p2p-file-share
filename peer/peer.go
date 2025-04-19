package peer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Peer struct {
	ListenAddr string
	peers map[string]bool
	mutex sync.Mutex
}

func NewPeer(listenAddr string) *Peer {
	return &Peer{
		ListenAddr: listenAddr,
		peers: make(map[string]bool),
	}
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
	fmt.Println("  list - List known peers")
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

		case "list":
			p.mutex.Lock()
			fmt.Println("Known peers:")
			for peer := range p.peers {
				fmt.Println(" -", peer)
			}
			p.mutex.Unlock()

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

		conn.Write([]byte("Received your message: " + message + "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
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

func (p *Peer) addPeer(address string){
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.peers[address] = true
	fmt.Println("Current peers:", p.peers)
}

