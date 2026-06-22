package clients

import (
	"fmt"
	"log/slog"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUserInfoClient(log *slog.Logger) (pb.UserInfoClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Error("error with creating grpc userInfo client", "error", err)
		return nil, nil, fmt.Errorf("failed to connect to grpc server: %w", err)
	}

	client := pb.NewUserInfoClient(conn)

	log.Info("grpc userInfo client successfully initialized 🚀🚀🚀")
	return client, conn, nil
}
