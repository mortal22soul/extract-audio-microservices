package clients

import (
	"fmt"
	"log"

	pb "github.com/video-converter/shared/proto/gen/go/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCClients manages all gRPC client connections
type GRPCClients struct {
	Auth          pb.AuthServiceClient
	authConn      *grpc.ClientConn
}

// NewGRPCClients creates real gRPC connections to backend services
func NewGRPCClients(authAddr, analyticsAddr string) (*GRPCClients, error) {
	// Connect to auth service
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service at %s: %w", authAddr, err)
	}

	authClient := pb.NewAuthServiceClient(authConn)
	log.Printf("Connected to auth service at %s", authAddr)

	return &GRPCClients{
		Auth:     authClient,
		authConn: authConn,
	}, nil
}

// Close closes all gRPC connections
func (c *GRPCClients) Close() {
	if c.authConn != nil {
		if err := c.authConn.Close(); err != nil {
			log.Printf("Failed to close auth gRPC connection: %v", err)
		}
	}
}