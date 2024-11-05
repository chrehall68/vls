package main

import (
	"github.com/chrehall68/vls/internal/vlsp"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopmentConfig().Build()

	// Start the server
	vlsp.StartServer(logger)
}
