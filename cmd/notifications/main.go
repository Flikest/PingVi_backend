package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
)

func main() {

	boot := rkboot.NewBoot()

	entry := rkgin.GetGinEntry("notifications")

	boot.AddShutdownHookFunc("shotdown-notifications-service", func() {
		fmt.Println("shotdown notificatons conects")
	})

	ctx := context.Background()

	boot.Bootstrap(ctx)

	boot.WaitForShutdownSig(ctx)
}
