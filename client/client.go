package client

import (
	"context"
	"log"
	"time"

	pb "github.com/shadowWorkflow/proto"
	"google.golang.org/grpc"
)

func Call(serverAddress string) {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewTasksClient(conn)

	// Create a Payload
	payload := &pb.Payload{
		Data: []byte("Hello from client"),
	}

	// Call the gRPC method
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.Call(ctx, payload)
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	log.Printf("Response from server: %s", string(response.Data))
}
