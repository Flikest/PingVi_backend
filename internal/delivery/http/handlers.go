package deliveryhttp

import (
	restapi "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/gin-gonic/gin"
)

type HandlerSSO struct {
	Router  *gin.Engine
	Service *restapi.ServiceSSO
}

type HandlerMessenger struct {
	Router  *gin.Engine
	Service *restapi.ServiceMessenger
}

type HandlerSFU struct {
	Router  *gin.Engine
	Service *restapi.ServiceSFU
}

type HandlerFileStorage struct {
	Router  *gin.Engine
	Service *restapi.ServiceFileStorage
}

type HandlerBot struct {
	Router  *gin.Engine
	Service *restapi.ServiceBots
}

func RegisterSSORouter(h *HandlerSSO) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		userRouter := v1.Group("/users")
		{
			userRouter.GET("/:id", h.Service.SelectUserByID)
			userRouter.GET("/me", h.Service.GetMeProfile)
			userRouter.GET("/sessions", h.Service.SelectUserSessions)
			userRouter.POST("/refresh", h.Service.RefreshTokens)
			userRouter.POST("/logup", h.Service.Logup)
			userRouter.POST("/login", h.Service.Login)
			userRouter.POST("/logout", h.Service.Logout)
			userRouter.POST("/2fa", h.Service.Verify2FA)
			userRouter.POST("/recover_password", h.Service.RecoverPassword)
			userRouter.POST("/temporary_link", h.Service.SendTemporaryLink)
			userRouter.PUT("/", h.Service.UpdateUser)
			userRouter.PATCH("/password", h.Service.UpdatePassword)
			userRouter.PATCH("/enable_2fa", h.Service.EnableTwoFactorAuth)
			userRouter.PATCH("/disable_2fa", h.Service.DisableTwoFactorAuth)
			userRouter.DELETE("/", h.Service.DeleteUser)
		}
	}

	return h.Router
}

func RegisterMessengerRouter(h *HandlerMessenger) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		messageRouter := v1.Group("/ws")
		{
			messageRouter.GET("/:session_id", h.Service.Handshake)
			messageRouter.GET("/preview/:link", h.Service.LinkPreview)
			messageRouter.GET("/messages/:chat_id", h.Service.GetAllMessageFromChat)
			messageRouter.DELETE("/messages/clear", h.Service.ClearMesagesFromChat)
		}

		reactionRouter := v1.Group("reactions")
		{
			reactionRouter.GET("/reactions/:chat_id", h.Service.GetAllReactionsFromChat)
		}

		channelRouter := v1.Group("/channels")
		{
			channelRouter.GET("/join/:channel_id/", h.Service.JoinChannel)
			channelRouter.GET("/cick", h.Service.KickChannelMember)
			channelRouter.GET("/leave/:channel_id/", h.Service.LeaveFromChannel)
			channelRouter.GET("/:user_id", h.Service.GetChannels)
			channelRouter.POST("/", h.Service.CreateChannel)
			channelRouter.PUT("/", h.Service.UpdateChannel)
			channelRouter.DELETE("/", h.Service.Deletechannel)
		}

		groupRouter := v1.Group("/groups")
		{
			groupRouter.GET("/join/:group_id/", h.Service.JoinGroup)
			groupRouter.GET("/leave/:group_id/", h.Service.LeaveGroup)
			groupRouter.GET("/member/all/:group_id", h.Service.SelectAllMembersGroup)
			groupRouter.GET("member/:group_id", h.Service.SelectMemberGroup)
			groupRouter.GET("/", h.Service.SelectGroup)
			groupRouter.POST("/", h.Service.CreateGroup)
			groupRouter.PUT("/", h.Service.UpdateGroup)
			groupRouter.DELETE("/kick", h.Service.KickMemberFromGroup)
			groupRouter.DELETE("/", h.Service.DeleteGroup)
		}

		personalChatRouter := v1.Group("/personal_chat")
		{
			personalChatRouter.POST("/", h.Service.CreatePersonalChat)
			personalChatRouter.DELETE("/", h.Service.DeletePersonalChat)
		}

		roleRouter := v1.Group("/roles")
		{
			roleRouter.GET("/permissions/:channel_id/:user_id", h.Service.GetPermissions)
			roleRouter.POST("/", h.Service.CreateRole)
			roleRouter.GET("/:channel_id", h.Service.GetAllRoles)
			roleRouter.PUT("/", h.Service.UpdateRole)
			roleRouter.DELETE("/", h.Service.DeleteRole)
		}

		topicRouter := v1.Group("/topics")
		{
			topicRouter.POST("/", h.Service.CreateTopic)
			topicRouter.GET("/:channel_id", h.Service.GetAllTopics)
			topicRouter.PUT("/", h.Service.UpdateTopic)
			topicRouter.DELETE("/", h.Service.DeleteTopic)
		}

		directoryRouter := v1.Group("/directorys")
		{
			directoryRouter.GET("/:channel_id", h.Service.SelectAllDirectorys)
			directoryRouter.POST("/", h.Service.CreateDirectory)
			directoryRouter.PUT("/", h.Service.UpdateDirectory)
			directoryRouter.DELETE("/", h.Service.DeleteDirectory)
		}
	}

	return h.Router
}

func RegisterSFURouter(h *HandlerSFU) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		roomRouter := v1.Group("/rooms")
		{
			roomRouter.POST("/create_join_token", h.Service.CreateJoinToken)
		}
	}
	return h.Router
}

func RegisterFileStorageRouter(h *HandlerFileStorage) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		fileRouter := v1.Group("/files")
		{
			fileRouter.POST("/", h.Service.UpdateInFileStorage)
			fileRouter.PUT("/", h.Service.UpdateInFileStorage)
			fileRouter.POST("/temporary_url", h.Service.IssueTemporaryURL)
			fileRouter.DELETE("/", h.Service.DeleteFromFileStorage)
		}
	}

	return h.Router
}

func RegisterBotRouter(h *HandlerBot) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		botRouter := v1.Group("bots")
		{
			botRouter.POST("/create", h.Service.CreateBot)
			botRouter.GET("/updates", h.Service.GetMessages)
			botRouter.POST("/commands", h.Service.SendAnswers)
			botRouter.GET("/me", h.Service.GetMyBots)
			botRouter.GET("/:id", h.Service.GetBotById)
			botRouter.PUT("/update", h.Service.UpdateBot)
			botRouter.DELETE("/delete", h.Service.DeleteBot)
		}
	}

	return h.Router
}
