package main

import (
	"os"

	"github.com/chrehall68/vls/internal/vlsp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func logInit(debug bool, f *os.File) *zap.Logger {

	pe := zap.NewProductionEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeTime = zapcore.ISO8601TimeEncoder // The encoder can be customized for each output

	level := zap.InfoLevel
	if debug {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), level),
	)

	l := zap.New(core) //Creating the logger

	return l
}

func main() {
	outputFile, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	logger := logInit(true, outputFile)

	// Start the server
	vlsp.StartServer(logger)
}
