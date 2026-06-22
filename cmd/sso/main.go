package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Flikest/PingVi_backend/internal/config"
	"github.com/Flikest/PingVi_backend/internal/database/postgres"
	deliverygrpc "github.com/Flikest/PingVi_backend/internal/delivery/grpc"
	deliveryhttp "github.com/Flikest/PingVi_backend/internal/delivery/http"
	"github.com/Flikest/PingVi_backend/internal/repository"
	servicegrpc "github.com/Flikest/PingVi_backend/internal/services/grpc"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/Flikest/PingVi_backend/pkg/logger"
	"github.com/joho/godotenv"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
)

// @title PingVi is a broker who will take care of you.
// @version 1.0
// @description This is an API for interacting with the PingVi broker.
// @termsOfService http://swagger.io/terms/

// @securityDefinitions.basic BasicAuth

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	env := flag.String("env", "local", "enviroment variable")
	flag.Parse()

	log := logger.NewLogger(*env)

	_, err := config.ParseConfig(log, fmt.Sprintf("./config/sso/sso.config.%s.yaml", *env))
	if err != nil {
		log.Error("error with parsing yaml config: ", "error", err)
		panic("failed to get records from config file")
	}

	if err := godotenv.Load(); err != nil {
		log.Error("error loading environment: ", "error", err)
	}

	boot := rkboot.NewBoot()

	ginEntry := rkgin.GetGinEntry("sso")
	grpcEntry := rkgrpc.GetGrpcEntry("user_info")

	dbCtx := context.Background()

	db := postgres.MustPostgresDBOpen(&postgres.PostgresConig{
		Ctx:      dbCtx,
		ConnPath: os.Getenv("POSTGRES_CONNECTION_PATH"),
	})

	ssoRepository := repository.NewRepositorySSO(&repository.RepositorySSO{
		Log: log,
		DB:  db,
	})

	ssoService := servicehttp.NewSSOService(&servicehttp.ServiceSSO{
		Log:        log,
		Repository: ssoRepository,
	})

	_ = deliveryhttp.RegisterSSORouter(&deliveryhttp.HandlerSSO{
		Router:  ginEntry.Router,
		Service: ssoService,
	})

	boot.AddShutdownHookFunc("close-db-connection", func() {
		fmt.Println("closing connections to the database")

		// TODO: Logic for closing connections to the database
	})

	var userInfoRepository servicegrpc.UserInfoRepository = ssoRepository

	userInfoService := servicegrpc.NewServiceUserInfo(log, userInfoRepository)

	deliverygrpc.RegisterUserInfoServer(grpcEntry.Server, userInfoService, log)

	boot.Bootstrap(context.Background())

	boot.WaitForShutdownSig(context.Background())
}
