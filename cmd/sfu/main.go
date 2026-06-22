package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/Flikest/PingVi_backend/internal/clients"
	deliveryhttp "github.com/Flikest/PingVi_backend/internal/delivery/http"
	servicehttp "github.com/Flikest/PingVi_backend/internal/services/http"
	"github.com/Flikest/PingVi_backend/pkg/logger"
	"github.com/joho/godotenv"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

func main() {
	env := flag.String("env", "local", "enviroment variable")

	log := logger.NewLogger(*env)

	if err := godotenv.Load(fmt.Sprintf("sfu.%s.env", *env)); err != nil {
		log.Error("error loading environment: ", "error", err)
	}

	boot := rkboot.NewBoot()

	ginEntry := rkgin.GetGinEntry("sfu")

	client, conn, err := clients.NewPermissionsClient(log)
	if err != nil {
		log.Error("error with creating permission client: ", "error", err)
	}

	sfuService := servicehttp.NewSFUService(&servicehttp.ServiceSFU{
		Log:    log,
		Client: client,
	})

	_ = deliveryhttp.RegisterSFURouter(&deliveryhttp.HandlerSFU{
		Router:  ginEntry.Router,
		Service: sfuService,
	})

	boot.AddShutdownHookFunc("shotdown-sfu-service", func() {
		log.Info("shutdown sfu service ⏻⏻⏻")
		conn.Close()
	})

	boot.Bootstrap(context.Background())
	boot.WaitForShutdownSig(context.Background())
}
