package servicehttp

import (
	"net/http"
	"os"
	"time"

	pb "github.com/Flikest/PingVi_backend/gen/go/permissions"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
)

func (s *ServiceSFU) CreateJoinToken(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.CreateJoinToken
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error(": ", "error", err)
		ctx.JSON(http.StatusBadRequest, "invalid body")
		return
	}

	body.UserID = userID

	permission, err := s.GrpcClient.GetUserPermission(ctx.Request.Context(), &pb.GetUserPermissionRequest{
		UserId: body.UserID.String(),
	})
	if err != nil {
		s.Log.Error("error with selectnig permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if permission.Permission[6] != '1' {
		ctx.JSON(http.StatusForbidden, "not enough permissions")
		return
	}

	at := auth.NewAccessToken(os.Getenv("LIIVE_KIT_API_KEY"), os.Getenv("LIVE_KIT_API_SECRET"))
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     body.TopicID.String(),
	}
	at.SetVideoGrant(grant).
		SetIdentity(body.UserID.String()).
		SetValidFor(time.Hour)

	joinToken, err := at.ToJWT()
	if err != nil {
		s.Log.Error("error with JWT token generation: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, joinToken)
}
