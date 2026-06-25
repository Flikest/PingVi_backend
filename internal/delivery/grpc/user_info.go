package deliverygrpc

import (
	"context"
	"log/slog"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	servicegrpc "github.com/Flikest/PingVi_backend/internal/services/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandlerUserInfo struct {
	pb.UnimplementedUserInfoServer
	Service *servicegrpc.ServiceUserInfo
	Log     *slog.Logger
}

func RegisterUserInfoServer(
	srv *grpc.Server,
	service *servicegrpc.ServiceUserInfo,
	log *slog.Logger,
) {
	pb.RegisterUserInfoServer(srv, &HandlerUserInfo{
		Service: service,
		Log:     log,
	})
}

func (h *HandlerUserInfo) GetUserNameByID(
	ctx context.Context,
	req *pb.GetUserNameByIdRequest,
) (*pb.GetUserNameByIdResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		h.Log.Error("error with parsing user id: ", "error", err)
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	name, err := h.Service.GetUserNameByID(ctx, userID)
	if err != nil {
		h.Log.Error("error with selecting user name: ", "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserNameByIdResponse{
		Name: name,
	}, nil
}

func (h *HandlerUserInfo) GetUserIDBySessionID(
	ctx context.Context,
	req *pb.GetUserIDBySessionIDRequest,
) (*pb.GetUserIDBySessionIDResponse, error) {
	sessionID := req.GetSessionId()

	if sessionID == "" {
		h.Log.Error("error with parsing session id")
		return nil, status.Error(codes.InvalidArgument, "invalid session id")
	}

	userID, err := h.Service.GetUserIDBySessionID(ctx, sessionID)
	if err != nil {
		h.Log.Error("error with selecting user id: ", "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetUserIDBySessionIDResponse{
		UserId: userID.String(),
	}, nil
}
