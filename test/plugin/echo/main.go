package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
)

// ListenPort is the port that the echo server is listening on
const ListenPort = "6174"

func main() {
	log.Printf("Listening on %s\n", ListenPort)
	server, err := net.Listen("tcp", ":"+ListenPort)
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	defer server.Close()

	connectionChannel := acceptConnections(server)
	for {
		go handleConnnection(<-connectionChannel)
	}
}

func acceptConnections(listener net.Listener) chan net.Conn {
	connectionChannel := make(chan net.Conn)
	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				log.Printf("Could not accept connection: " + err.Error())
				continue
			}

			log.Printf("%v connection to %v opened.\n", client.RemoteAddr(), client.LocalAddr())

			connectionChannel <- client
		}
	}()

	return connectionChannel
}

func handleConnnection(client net.Conn) {
	bufferedReader := bufio.NewReader(client)
	for {
		line, err := bufferedReader.ReadBytes('\n')
		if err != nil {
			log.Println("Connection closed by client")
			client.Close()
			return
		}

		if bytes.Equal(line, []byte("\r\n")) {
			log.Println("Connection closed by echo server")
			client.Close()
			return
		}

		client.Write(line)
	}
}
