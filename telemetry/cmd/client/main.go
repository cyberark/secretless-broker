package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/cyberark/secretless-broker/internal/plugin/connectors/tcp/mssql"
)


func clientConnect(address string) (net.Conn, error) {
	connector, err := mssql.NewMSSQLConnector("sqlserver://" + address)
	if err != nil {
		return nil, err
	}

	return connector.Connect(context.Background())
}

func randBuf(size int) []byte {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		log.Printf("randBuf error: %v", err)
	}
	return buf
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide host:port.")
		return
	}

	c, err := clientConnect(arguments[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	readSize := 16
	writeSize := 16

	go func() {
		for {
			received := make([]byte, readSize)
			n, err := c.Read(received)
			if err == io.EOF {
				return
			}
			fmt.Println("received:", serialize(received[:n]))
		}
	}()
	for {
		sending := randBuf(writeSize)
		fmt.Println("sending:", serialize(sending))
		_, err = c.Write(sending)
		if err == io.EOF {
			fmt.Println("connection closed");
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func serialize(somebytes []byte) string {
	return hex.EncodeToString(somebytes)
}