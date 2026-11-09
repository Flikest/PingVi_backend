package deliveryhttp

import (
	"github.com/Flikest/PingVi_backend/internal/middleware"
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

type HandlerExhcnage struct {
	Router  *gin.Engine
	Service *restapi.ServiceWallet
}

func RegisterSSORouter(h *HandlerSSO) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		userRouter := v1.Group("/users")
		{
			userRouter.GET("/:id", middleware.IsAuthorized, h.Service.SelectUserByID)
			userRouter.GET("/me", middleware.IsAuthorized, h.Service.GetMeProfile)
			userRouter.GET("/sessions", middleware.IsAuthorized, h.Service.SelectUserSessions)
			userRouter.POST("/refresh", h.Service.RefreshTokens)
			userRouter.POST("/logup", h.Service.Logup)
			userRouter.POST("/login", h.Service.Login)
			userRouter.POST("/logout", middleware.IsAuthorized, h.Service.Logout)
			userRouter.POST("/2fa", middleware.IsAuthorized, h.Service.Verify2FA)
			userRouter.POST("/recover_password", middleware.IsAuthorized, h.Service.RecoverPassword)
			userRouter.POST("/temporary_link", middleware.IsAuthorized, h.Service.SendTemporaryLink)
			userRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateUser)
			userRouter.PATCH("/password", middleware.IsAuthorized, h.Service.UpdatePassword)
			userRouter.PATCH("/enable_2fa", middleware.IsAuthorized, h.Service.EnableTwoFactorAuth)
			userRouter.PATCH("/disable_2fa", middleware.IsAuthorized, h.Service.DisableTwoFactorAuth)
			userRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteUser)
			userRouter.DELETE("/session/:session_id", middleware.IsAuthorized, h.Service.DeleteSession)
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
			messageRouter.GET("/preview/:link", middleware.IsAuthorized, h.Service.LinkPreview)
			messageRouter.GET("/messages/:chat_id", middleware.IsAuthorized, h.Service.GetAllMessageFromChat)
			messageRouter.DELETE("/messages/clear", middleware.IsAuthorized, h.Service.ClearMesagesFromChat)
		}

		chatsRouter := v1.Group("/chats")
		{
			chatsRouter.GET("/me", middleware.IsAuthorized, h.Service.GetStartMyChats)
		}

		reactionRouter := v1.Group("reactions")
		{
			reactionRouter.GET("/reactions/:chat_id", middleware.IsAuthorized, h.Service.GetAllReactionsFromChat)
		}

		channelRouter := v1.Group("/channels")
		{
			channelRouter.GET("/join/:channel_id/", middleware.IsAuthorized, h.Service.JoinChannel)
			channelRouter.POST("/kick", middleware.IsAuthorized, h.Service.KickChannelMember)
			channelRouter.GET("/leave/:channel_id/", middleware.IsAuthorized, h.Service.LeaveFromChannel)
			channelRouter.GET("/member/all/:channel_id", middleware.IsAuthorized, h.Service.GetChannelMembers)
			channelRouter.GET("/:user_id", middleware.IsAuthorized, h.Service.GetChannels)
			channelRouter.POST("/", middleware.IsAuthorized, h.Service.CreateChannel)
			channelRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateChannel)
			channelRouter.DELETE("/", middleware.IsAuthorized, h.Service.Deletechannel)
		}

		groupRouter := v1.Group("/groups")
		{
			groupRouter.GET("/join/:group_id/", middleware.IsAuthorized, h.Service.JoinGroup)
			groupRouter.GET("/leave/:group_id/", middleware.IsAuthorized, h.Service.LeaveGroup)
			groupRouter.GET("/member/all/:group_id", middleware.IsAuthorized, h.Service.SelectAllMembersGroup)
			groupRouter.GET("member/:group_id", middleware.IsAuthorized, h.Service.SelectMemberGroup)
			groupRouter.GET("/", middleware.IsAuthorized, h.Service.SelectGroup)
			groupRouter.POST("/", middleware.IsAuthorized, h.Service.CreateGroup)
			groupRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateGroup)
			groupRouter.DELETE("/kick", middleware.IsAuthorized, h.Service.KickMemberFromGroup)
			groupRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteGroup)
		}

		personalChatRouter := v1.Group("/personal_chat")
		{
			personalChatRouter.POST("/", middleware.IsAuthorized, h.Service.CreatePersonalChat)
			personalChatRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeletePersonalChat)
		}

		roleRouter := v1.Group("/roles")
		{
			roleRouter.GET("/permissions/:channel_id/", middleware.IsAuthorized, h.Service.GetPermissions)
			roleRouter.POST("/", middleware.IsAuthorized, h.Service.CreateRole)
			roleRouter.GET("/:channel_id", middleware.IsAuthorized, h.Service.GetAllRoles)
			roleRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateRole)
			roleRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteRole)
		}

		topicRouter := v1.Group("/topics")
		{
			topicRouter.POST("/", middleware.IsAuthorized, h.Service.CreateTopic)
			topicRouter.GET("/:channel_id", middleware.IsAuthorized, h.Service.GetAllTopics)
			topicRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateTopic)
			topicRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteTopic)
		}

		directoryRouter := v1.Group("/directorys")
		{
			directoryRouter.GET("/:channel_id", middleware.IsAuthorized, h.Service.SelectAllDirectorys)
			directoryRouter.POST("/", middleware.IsAuthorized, h.Service.CreateDirectory)
			directoryRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateDirectory)
			directoryRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteDirectory)
		}
	}

	return h.Router
}

func RegisterSFURouter(h *HandlerSFU) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		roomRouter := v1.Group("/rooms")
		{
			roomRouter.POST("/create_join_token", middleware.IsAuthorized, h.Service.CreateJoinToken)
		}
	}
	return h.Router
}

func RegisterFileStorageRouter(h *HandlerFileStorage) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		fileRouter := v1.Group("/files")
		{
			fileRouter.POST("/", middleware.IsAuthorized, h.Service.UploadToFileStorage)
			fileRouter.PUT("/", middleware.IsAuthorized, h.Service.UpdateInFileStorage)
			fileRouter.POST("/temporary_url", middleware.IsAuthorized, h.Service.IssueTemporaryURL)
			fileRouter.DELETE("/", middleware.IsAuthorized, h.Service.DeleteFromFileStorage)
		}
	}

	return h.Router
}

func RegisterBotRouter(h *HandlerBot) *gin.Engine {
	v1 := h.Router.Group("/v1")
	{
		botRouter := v1.Group("bots")
		{
			botRouter.POST("/create", middleware.IsAuthorized, h.Service.CreateBot)
			botRouter.GET("/updates", middleware.IsAuthorized, h.Service.GetMessages)
			botRouter.POST("/answers", middleware.IsAuthorized, h.Service.SendAnswers)
			botRouter.GET("/me", middleware.IsAuthorized, h.Service.GetMyBots)
			botRouter.GET("/:id", middleware.IsAuthorized, h.Service.GetBotById)
			botRouter.PUT("/update", middleware.IsAuthorized, h.Service.UpdateBot)
			botRouter.DELETE("/delete/:bot_id", middleware.IsAuthorized, h.Service.DeleteBot)
		}

		commandsRouter := v1.Group("commands")
		{
			commandsRouter.POST("/create", middleware.IsAuthorized, h.Service.CreateBotCommand)
			commandsRouter.GET("/:bot_id", middleware.IsAuthorized, h.Service.GetBotCommands)
			commandsRouter.PUT("/update", middleware.IsAuthorized, h.Service.UpdateBotCommand)
			commandsRouter.DELETE("/delete/:bot_id/:command_id", middleware.IsAuthorized, h.Service.DeleteBotCommand)
		}
	}

	return h.Router
}

func RegisterExchangeRouter(h *HandlerExhcnage) {
	v1 := h.Router.Group("/v1")
	{
		exchangeRouter := v1.Group("/exchange")
		{
			exchangeRouter.GET("/balance/:account_type", middleware.IsAuthorized, h.Service.GetBalance)
			exchangeRouter.GET("/transactions", middleware.IsAuthorized, h.Service.GetTransactions)
			exchangeRouter.POST("/create", middleware.IsAuthorized, h.Service.CreateWallet)
			exchangeRouter.GET("/snapshots", middleware.IsAuthorized, h.Service.GetSnapshot)
			exchangeRouter.POST("/transaction/send", middleware.IsAuthorized, h.Service.SendTransaction)
			exchangeRouter.GET("/swop/list_currencies_conversion/:exchange_name", middleware.IsAuthorized, h.Service.GetListCurrenciesConversion)
			exchangeRouter.POST("/swop/convert_quote", middleware.IsAuthorized, h.Service.GetConvertQuote)
			exchangeRouter.POST("/swop/convert", middleware.IsAuthorized, h.Service.ConfirmConvertQuote)
			exchangeRouter.POST("/ws", middleware.IsAuthorized, h.Service.Handshake)
			exchangeRouter.GET("/instruments/:instrument", middleware.IsAuthorized, h.Service.Searchinstruments)
			exchangeRouter.GET("/canldes/:instrument/:exchange/:interval/:after", middleware.IsAuthorized, h.Service.GetCandles)
			exchangeRouter.GET("/order_book/:instrument/:exchange/:interval/:after", middleware.IsAuthorized, h.Service.GetOrderBook)
			exchangeRouter.POST("/api_keys/add", middleware.IsAuthorized, h.Service.AddExchangeApiKeys)
			exchangeRouter.PUT("/api_keys/update", middleware.IsAuthorized, h.Service.UpdateExchangeApiKeys)
			exchangeRouter.DELETE("/api_keys/delete", middleware.IsAuthorized, h.Service.DeleteExchangeApiKeys)
		}
	}
}
