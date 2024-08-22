package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/shadowWorkflow/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrMissingAddress = errors.New("missing address")
)

// Config is the configuration for a TestTasks.
type Config struct {
	Addr      string      // Addr is the address to listen on (e.g. ":8080")
	TLSConfig *tls.Config // TLSConfig is the TLS configuration to use for the gRPC server
	TCPAddr   string      // TCPAddr is the address for the TCP connection (e.g. "127.0.0.1")
	TCPPort   string      // TCPPort is the port for the TCP connection (e.g. "6000")
	EnableTLS bool        // EnableTLS indicates if TLS should be used for the TCP connection
}

// TestTasks is a gRPC server that implements the TasksServer interface.
type TestTasks struct {
	pb.UnimplementedTasksServer // Embed the UnimplementedTasksServer for forward compatibility
	sync.RWMutex

	cfg     Config       // cfg is the user-provided configuration
	ln      net.Listener // ln is the net listener
	server  *grpc.Server // server is the gRPC server
	logger  *log.Logger  // logger is used to log information
	tcpConn net.Conn     // tcpConn is the TCP connection object
}

// New creates a new TestTasks with the provided configuration.
func New(cfg Config, logger *log.Logger) (*TestTasks, error) {
	tp := &TestTasks{
		cfg:    cfg,
		logger: logger,
	}

	// Verify minimum configuration
	if tp.cfg.Addr == "" {
		return nil, ErrMissingAddress
	}

	// Create a net listener
	ln, err := net.Listen("tcp", tp.cfg.Addr)
	if err != nil {
		return nil, err
	}
	tp.ln = ln

	// Set TLS if provided, otherwise use insecure
	creds := insecure.NewCredentials()
	if tp.cfg.TLSConfig != nil {
		creds = credentials.NewTLS(tp.cfg.TLSConfig)
	}

	// Create the gRPC server
	tp.server = grpc.NewServer(grpc.Creds(creds))

	// Register the gRPC server
	pb.RegisterTasksServer(tp.server, tp)

	// Initialize TCP connection
	err = tp.initTCPConnection()
	if err != nil {
		return nil, err
	}

	return tp, nil
}

// initTCPConnection initializes the TCP connection as per the configuration.
func (tp *TestTasks) initTCPConnection() error {
	address := fmt.Sprintf("%s:%s", tp.cfg.TCPAddr, tp.cfg.TCPPort)
	var conn net.Conn
	var err error

	if tp.cfg.EnableTLS {
		conn, err = tls.Dial("tcp", address, tp.cfg.TLSConfig)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		tp.logger.Printf("Failed to establish TCP connection: %v", err)
		return err
	}

	tp.tcpConn = conn
	tp.logger.Println("TCP connection established")
	return nil
}

// Start starts the gRPC server and blocks until the server is stopped.
func (tp *TestTasks) Start() error {
	tp.RLock()
	svr := tp.server
	ln := tp.ln
	tp.RUnlock()

	tp.logger.Println("Starting gRPC server...")
	defer tp.Shutdown()
	err := svr.Serve(ln)
	if err != nil && err != grpc.ErrServerStopped {
		tp.logger.Printf("Failed to start gRPC server: %v", err)
		return err
	}

	return nil
}

// Shutdown shuts down the gRPC server and closes the net listener.
func (tp *TestTasks) Shutdown() error {
	tp.RLock()
	defer tp.RUnlock()

	tp.logger.Println("Shutting down gRPC server...")

	// Stop the gRPC server
	if tp.server != nil {
		tp.server.Stop()
	}

	// Close the net listener
	if tp.ln != nil {
		tp.ln.Close()
	}

	// Close the TCP connection
	if tp.tcpConn != nil {
		tp.tcpConn.Close()
	}

	tp.logger.Println("Shutdown complete")
	return nil
}

// Call implements the Call rpc call and returns the payload to the caller, effectively creating an echo server.
// It also strips the Envoy header and extracts the MLI.
func (tp *TestTasks) Call(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
	tp.logger.Println("Received gRPC request")

	// Simulate stripping the header and extracting MLI
	if len(req.Data) < 14 {
		tp.logger.Println("Invalid payload: insufficient data")
		return nil, errors.New("invalid payload: insufficient data")
	}

	header := req.Data[:10]
	mli := req.Data[10:14]

	tp.logger.Printf("Stripped header: %x, Extracted MLI: %x", header, mli)

	// Example: Send the stripped data over the TCP connection
	_, err := tp.tcpConn.Write(req.Data[14:])
	if err != nil {
		tp.logger.Printf("Failed to send data over TCP: %v", err)
		return nil, err
	}

	// Read the response from TCP server
	buffer := make([]byte, 1024)
	n, err := tp.tcpConn.Read(buffer)
	if err != nil {
		tp.logger.Printf("Failed to read response from TCP: %v", err)
		return nil, err
	}

	// Return the response as the gRPC response payload
	tp.logger.Println("Successfully processed gRPC request")
	return &pb.Payload{Data: buffer[:n]}, nil
}
