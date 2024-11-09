package main

import (
	"flag"
	"net"
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
	listenAt := flag.String("listen-at", "", "If specified (not empty string), listen at this address for a TCP connection instead of talking over stdio.")
	flag.Parse()

	if *listenAt != "" {
		logger := logInit(true, os.Stdout)
		sweet := logger.Sugar()

		listen, err := net.Listen("tcp", *listenAt)
		if err != nil {
			// Use %v to get a human facing error message
			sweet.Fatalf("listen-at '%s': %v", *listenAt, err)
			os.Exit(-1)
		}

		for {
			conn, err := listen.Accept()
			if err != nil {
				sweet.Errorf("listen-at: accept: %v", err)
			}

			vlsp.StartServer(logger, conn, conn)
		}
	} else {
		outputFile, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		logger := logInit(true, outputFile)

		vlsp.StartServer(logger, os.Stdin, os.Stdout)
	}
}
