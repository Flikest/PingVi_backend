package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Flikest/PingVi_backend/internal/clients"
	"github.com/Flikest/PingVi_backend/internal/config"
	"github.com/Flikest/PingVi_backend/internal/database/redis"
	"github.com/Flikest/PingVi_backend/internal/database/scylla"
	deliveryhttp "github.com/Flikest/PingVi_backend/internal/delivery/http"
	"github.com/Flikest/PingVi_backend/internal/repository"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/Flikest/PingVi_backend/pkg/logger"
	"github.com/bwmarrin/snowflake"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

//	@title			PingVi messenger
//	@version		1.0
//	@description	This is PingVi messenger.
//	@termsOfService	http://swagger.io/terms/

//	@Host		localhost:8082
//	@BasePath	/v1

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	env := flag.String("env", "local", "enviroment variable")

	nodeNumber := flag.Int64("node-number", 1, "node number")

	flag.Parse()

	log := logger.NewLogger(*env)

	_, err := config.ParseConfig(log, *env, "./configs/messenger/messenger.yaml")
	if err != nil {
		panic(fmt.Errorf("not parse config: %w", err))
	}

	boot := rkboot.NewBoot(rkboot.WithBootConfigPath("./configs/messenger/messenger.yaml", nil))

	entry := rkgin.GetGinEntry("messenger")

	ctx := context.Background()

	scyllaSession := scylla.MustScyllaDBOpen()

	redisClient := redis.NewRedisCleint()

	repository := repository.NewRepositoryMessenger(&repository.RepositoryMessenger{
		Log:     log,
		Session: scyllaSession,
	})

	node, err := snowflake.NewNode(*nodeNumber)
	if err != nil {
		fmt.Println(err)
		return
	}

	hub := servicehttp.NewHub(&servicehttp.Hub{
		Clients:             make(map[string]*servicehttp.Client),
		Redis:               redisClient,
		RepositoryMessenger: repository,
		Log:                 log,
		Node:                node,
	})

	client, conn, err := clients.NewUserInfoClient(log)
	if err != nil {
		panic(fmt.Errorf("failed to connect to userinfo service: %w", err))
	}

	service := servicehttp.NewMessengerService(&servicehttp.ServiceMessenger{
		Log:        log,
		Repository: repository,
		Client:     client,
		Hub:        hub,
	})

	_ = deliveryhttp.RegisterMessengerRouter(&deliveryhttp.HandlerMessenger{
		Router:  entry.Router,
		Service: service,
	})

	boot.AddShutdownHookFunc("shotdown-messenger-service", func() {
		log.Info("shutdown messenger service")
		conn.Close()
		if redisClient != nil {
			redisClient.Close()
		}
	})
	boot.Bootstrap(ctx)

	boot.WaitForShutdownSig(ctx)
}
