package main

import (
	"flag"
	"fmt"

	"github.com/Flikest/PingVi_backend/internal/config"
	"github.com/Flikest/PingVi_backend/pkg/logger"
)

func main() {
	env := flag.String("env", "local", "enviroment variable")
	flag.Parse()

	log := logger.NewLogger(*env)

	_, err := config.ParseConfig(log, *env)
	if err != nil {
		panic(fmt.Sprintf("not parse config: %w", err))
	}

}
