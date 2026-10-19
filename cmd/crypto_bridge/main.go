package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Flikest/PingVi_backend/internal/database/postgres"
	deliveryhttp "github.com/Flikest/PingVi_backend/internal/delivery/http"
	"github.com/Flikest/PingVi_backend/internal/repository"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/Flikest/PingVi_backend/pkg/logger"
	"github.com/joho/godotenv"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

func main() {
	env := flag.String("env", "local", "enviroment variable")
	flag.Parse()

	log := logger.NewLogger(*env)

	if err := godotenv.Load(); err != nil {
		log.Warn("Error loading .env file", "error", err)
	}

	boot := rkboot.NewBoot(rkboot.WithBootConfigPath("./configs/crypto_bridge/crypto_bridge.yaml", nil))

	entry := rkgin.GetGinEntry("crypto_bridge")

	ctx := context.Background()

	db := postgres.MustPostgresDBOpen(&postgres.PostgresConig{
		Ctx:      ctx,
		ConnPath: os.Getenv("POSTGRES_CONNECTION_PATH"),
	})

	reposytory := repository.NewRepositoryWallet(&repository.RepositoryWallet{
		DB:  db,
		Log: log,
	})

	service := servicehttp.NewWalletService(&servicehttp.ServiceWallet{
		Repository: reposytory,
		Log:        log,
	})

	deliveryhttp.RegisterExchangeRouter(&deliveryhttp.HandlerExhcnage{
		Service: service,
		Router:  entry.Router,
	})

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
