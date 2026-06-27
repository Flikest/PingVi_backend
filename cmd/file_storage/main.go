package main

import (
	"context"
	"flag"

	"github.com/Flikest/PingVi_backend/internal/config"
	deliveryhttp "github.com/Flikest/PingVi_backend/internal/delivery/http"
	filestorage "github.com/Flikest/PingVi_backend/internal/file_storage"
	"github.com/Flikest/PingVi_backend/internal/repository"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/Flikest/PingVi_backend/pkg/logger"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

func main() {
	var env string

	flag.StringVar(&env, "env", "local", "environment variables")

	log := logger.NewLogger(env)

	config, err := config.ParseConfig(log, env, "./configs/file_storage/file_storage.yaml")
	if err != nil {
		panic("Failed to get minio config 😭😭😭")
	}

	boot := rkboot.NewBoot(rkboot.WithBootConfigPath("./configs/file_storage/file_storage.yaml", nil))

	entry := rkgin.GetGinEntry("file_storage")

	ctx := context.Background()

	minIOClient, err := filestorage.NewMinIOClient(ctx, filestorage.MinIOConfig{
		Endpoint: config.MinIOClient.Endpoint,
		UseSSL:   config.MinIOClient.UseSSL,
		Logger:   log,
	})

	repository := repository.NewReposossoryFileStorage(&repository.RepositoryFileStorage{
		Log:         log,
		MinIOClient: minIOClient,
	})

	service := servicehttp.NewFileStorageService(&servicehttp.ServiceFileStorage{
		Log:        log,
		Repository: repository,
	})

	_ = deliveryhttp.RegisterFileStorageRouter(&deliveryhttp.HandlerFileStorage{
		Router:  entry.Router,
		Service: service,
	})

	boot.AddShutdownHookFunc("shotdown-s3-service", func() {
		log.Info("shotdown notificatons conects")
	})
	boot.Bootstrap(ctx)

	boot.WaitForShutdownSig(ctx)
}
