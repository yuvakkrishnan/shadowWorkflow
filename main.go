package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/shadowWorkflow/client"
	"github.com/shadowWorkflow/logging"
	"github.com/shadowWorkflow/server"
)

func main() {
	// Initialize the logger
	logger := logging.InitializeLogger("WORKFLOW-1: ")

	// Start the TCP server
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener, err := net.Listen("tcp", "127.0.0.1:6000")
		if err != nil {
			fmt.Println("Error starting TCP server:", err)
			return
		}
		defer listener.Close()

		fmt.Println("TCP server listening on 127.0.0.1:6000")

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}

			go server.HandleConnection(conn)
		}
	}()

	// Wait for TCP server to start
	wg.Wait()

	// Define server configuration
	cfg := server.Config{
		Addr:    ":50051",
		TCPAddr: "127.0.0.1",
		TCPPort: "6000",
	}

	// Create new gRPC server with logging
	srv, err := server.New(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start gRPC server
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Call the client
	client.Call("localhost:50051")

	// Shutdown the server gracefully
	srv.Shutdown()
}
