package clients

import (
	"fmt"
	"log/slog"

	pb "github.com/Flikest/PingVi_backend/gen/go/permissions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewPermissionsClient(log *slog.Logger) (pb.PermissionsClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"localhost:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Error("error with creating grpc permissions client", "error", err)
		return nil, nil, fmt.Errorf("failed to connect to grpc server: %w", err)
	}

	client := pb.NewPermissionsClient(conn)

	log.Info("grpc permissions client successfully initialized 🚀🚀🚀")
	return client, conn, nil
}
