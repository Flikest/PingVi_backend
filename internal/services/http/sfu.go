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

// CreateJoinToken godoc
// @Summary      Create join token for room
// @Description  Creates a JWT token for joining a specific room. The token is valid for 1 hour and includes video grant permissions.
// @Description  Requires user authentication and the 'room_join' permission (permission[6] == '1').
// @Tags         sfu
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.CreateJoinToken  true  "Room join request"
// @Success      201  {string}  string  "JWT token for joining the room"  example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
// @Failure      400  {object}  map[string]interface{}  "Invalid request body or JWT token"  example("{\"error\":\"invalid body\"}")
// @Failure      403  {object}  map[string]interface{}  "User doesn't have enough permissions"  example("{\"error\":\"not enough permissions\"}")
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example("{\"error\":\"failed to generate JWT token\"}")
// @Router       /create_join_token [post]
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
