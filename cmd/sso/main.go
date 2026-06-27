package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Flikest/PingVi_backend/internal/config"
	"github.com/Flikest/PingVi_backend/internal/database/postgres"
	"github.com/Flikest/PingVi_backend/internal/database/redis"
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
	"google.golang.org/grpc"
)

func main() {
	env := flag.String("env", "local", "environment variable")
	flag.Parse()

	log := logger.NewLogger(*env)
	log.Info("Starting application", "env", *env)

	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file", "error", err)
	}

	_, err := config.ParseConfig(log, *env, "./configs/sso/sso.yaml")
	if err != nil {
		log.Error("Failed to parse config", "error", err)
		os.Exit(1)
	}
	log.Info("Configuration loaded successfully")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := postgres.MustPostgresDBOpen(&postgres.PostgresConig{
		Ctx:      ctx,
		ConnPath: os.Getenv("POSTGRES_CONNECTION_PATH"),
	})
	defer db.Close()
	log.Info("Database connection established")

	boot := rkboot.NewBoot(rkboot.WithBootConfigPath("./configs/sso/sso.yaml", nil))

	ginEntry := rkgin.GetGinEntry("sso")
	grpcEntry := rkgrpc.GetGrpcEntry("user_info")

	redisClient := redis.NewRedisCleint()

	ssoRepository := repository.NewRepositorySSO(&repository.RepositorySSO{
		Log: log,
		DB:  db,
		RDB: redisClient,
	})

	ssoService := servicehttp.NewSSOService(&servicehttp.ServiceSSO{
		Log:        log,
		Repository: ssoRepository,
	})

	userInfoRepository := ssoRepository

	userInfoService := servicegrpc.NewServiceUserInfo(log, userInfoRepository)

	grpcEntry.AddRegFuncGrpc(
		func(s *grpc.Server) {
			deliverygrpc.RegisterUserInfoServer(grpcEntry.Server, userInfoService, log)
		},
	)

	deliveryhttp.RegisterSSORouter(&deliveryhttp.HandlerSSO{
		Router:  ginEntry.Router,
		Service: ssoService,
	})

	log.Info("HTTP routes registered")

	log.Info("gRPC service registered")

	log.Info("All servers started successfully")

	boot.Bootstrap(ctx)

	boot.AddShutdownHookFunc("close-db-connection", func() {
		log.Info("Closing database connection")
		db.Close()
	})

	log.Info("Application started successfully")
	log.Info("HTTP server listening on http://localhost:8080")
	log.Info("gRPC server listening on port 50051")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Info("Context cancelled")
	case sig := <-sigChan:
		log.Info("Received signal", "signal", sig)
	}

	log.Info("Shutting down application...")
	boot.WaitForShutdownSig(ctx)
	log.Info("Application stopped")
}
