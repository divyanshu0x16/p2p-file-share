package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func startServer(port string){
	listener, err := net.Listen("tcp", ":"+port)

	if(err != nil ){
		fmt.Println("Error starting server:", err)
		return 
	}
	defer listener.Close()

	fmt.Println("Server listening on port", port)

	for {
		conn, err := listener.Accept()
		if(err != nil){
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn){
	defer conn.Close()
	fmt.Println("New connection from:", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		fmt.Println("Received:", message)

		conn.Write([]byte("Received your message: " + message + "\n"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

func connectToPeer(address string){
	conn, err := net.Dial("tcp", address)

	if ( err != nil ){
		fmt.Println("Error connecting to peer:", err)
	}

	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages to send (or 'exit' to quit):")

	for scanner.Scan(){
		message := scanner.Text()
		if message == "exit" {
			break
		}

		conn.Write([]byte(message + "\n"))

		responseReader := bufio.NewReader(conn)
		response, err := responseReader.ReadString('\n')
		if ( err != nil ){
			fmt.Println("Error reading response:", err)
			break
		}

		fmt.Println("Response:", strings.TrimSpace(response))
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run peer.go [server|client] [port|address]")
		fmt.Println("Example (server): go run peer.go server 8080")
		fmt.Println("Example (client): go run peer.go client 127.0.0.1:8080")
		return
	}
	
	mode := os.Args[1]
	arg := os.Args[2]
	
	switch mode {
	case "server":
		startServer(arg)
	case "client":
		connectToPeer(arg)
	default:
		fmt.Println("Invalid mode. Use 'server' or 'client'")
	}
}