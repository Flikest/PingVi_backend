package deliverygrpc

import (
	"context"
	"log/slog"

	pb "github.com/Flikest/PingVi_backend/gen/go/permissions"
	servicegrpc "github.com/Flikest/PingVi_backend/internal/services/grpc"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type HandlerPermissions struct {
	pb.UnimplementedPermissionsServer
	Service *servicegrpc.ServicePermissions
	Log     *slog.Logger
}

func RegisterPermissionsServer(
	srv *grpc.Server,
	service *servicegrpc.ServicePermissions,
	log *slog.Logger,
) {
	pb.RegisterPermissionsServer(srv, &HandlerPermissions{
		Service: service,
		Log:     log,
	})
}

func (h *HandlerPermissions) GetUserPermission(
	ctx context.Context,
	req *pb.GetUserPermissionRequest,
) (*pb.GetUserPermissionResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		h.Log.Error("Error parsing user ID", "error", err)
		return nil, err
	}

	permission, err := h.Service.GetUserPermission(ctx, userID)
	if err != nil {
		h.Log.Error("Error getting user permission", "error", err)
		return nil, err
	}

	h.Log.Info("GetUserPermission request completed")
	return &pb.GetUserPermissionResponse{
		Permission: permission,
	}, nil
}
