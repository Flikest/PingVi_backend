package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Flikest/PingVi_backend/pkg/logger"
	"github.com/joho/godotenv"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

func main() {
	var env string

	flag.StringVar(&env, "env", "local", "environment variables")

	if err := godotenv.Load(fmt.Sprintf("minIO.%s.env", env)); err != nil {
		panic(fmt.Sprintf("failed to get environment variables: %w", err))
	}

	log := logger.NewLogger(env)

	boot := rkboot.NewBoot()

	entry := rkgin.GetGinEntry("notifications")

	boot.AddShutdownHookFunc("shotdown-s3-service", func() {
		log.Info("shotdown notificatons conects")
	})

	ctx := context.Background()

	boot.Bootstrap(ctx)

	boot.WaitForShutdownSig(ctx)
}
