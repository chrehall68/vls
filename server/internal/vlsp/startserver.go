package vlsp

import (
	"context"
	"io"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// StartServer starts the language server.
// It reads from stdin and writes to stdout.
func StartServer(logger *zap.Logger, reader io.ReadCloser, writer io.WriteCloser) {
	stream := jsonrpc2.NewStream(&readWriteCloser{
		reader: reader,
		writer: writer,
	})
	conn := jsonrpc2.NewConn(stream)

	handler, ctx, err := NewHandler(
		context.Background(),
		protocol.ServerDispatcher(conn, logger),
		protocol.ClientDispatcher(conn, logger),
		logger,
		&stream,
	)

	if err != nil {
		logger.Sugar().Fatalf("while initializing handler: %w", err)
	}
	logger.Sugar().Infof("initialized handler")

	conn.Go(ctx, protocol.ServerHandler(
		handler, jsonrpc2.MethodNotFoundHandler,
	))
	<-conn.Done()
}

type readWriteCloser struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func (r *readWriteCloser) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)
	return n, err
}

func (r *readWriteCloser) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *readWriteCloser) Close() error {
	return multierr.Append(r.reader.Close(), r.writer.Close())
}
