package main

import (
	"flag"
	"net"
	"os"

	"github.com/chrehall68/vls/internal/vlsp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func logConfig() zapcore.EncoderConfig {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder // The encoder can be customized for each output
	return pe
}

func structuredLogCore(level zapcore.Level, f *os.File) zapcore.Core {
	config := logConfig()
	encoder := zapcore.NewJSONEncoder(config)

	return zapcore.NewCore(encoder, zapcore.AddSync(f), level)
}

func consoleLogCore(level zapcore.Level) zapcore.Core {
	config := logConfig()
	encoder := zapcore.NewConsoleEncoder(config)

	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

func main() {
	logToFile := flag.String("log-file", "", "If specified, output JSON structured log to this file.")
	logToConsole := flag.Bool("log-console", false, "If true and --listen-at is specified, output console log.")
	listenAt := flag.String("listen-at", "", "If specified (not empty string), listen at this address for a TCP connection instead of talking over stdio.")
	flag.Parse()

	level := zapcore.DebugLevel

	cores := []zapcore.Core{}
	if *logToConsole && *listenAt != "" {
		cores = append(cores, consoleLogCore(level))
	}
	if *logToFile != "" {
		outputFile, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		cores = append(cores, structuredLogCore(level, outputFile))
	}

	logger := zap.New(zapcore.NewTee(cores...))
	sugar := logger.Sugar()

	if *listenAt != "" {
		listen, err := net.Listen("tcp", *listenAt)
		if err != nil {
			// Use %v to get a human facing error message
			sugar.Fatalf("listen-at '%s': %v", *listenAt, err)
			os.Exit(-1)
		}

		for {
			conn, err := listen.Accept()
			if err != nil {
				sugar.Errorf("listen-at: accept: %v", err)
			}

			vlsp.StartServer(logger, conn, conn)
		}
	} else {
		vlsp.StartServer(logger, os.Stdin, os.Stdout)
	}
}
