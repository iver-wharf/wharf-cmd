package workerhttpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iver-wharf/wharf-core/pkg/ginutil"
)

// Server lets us start an HTTP server asynchronously.
type Server struct {
	address             string
	port                string
	onServeErrorHandler func(error)

	srv       *http.Server
	isRunning bool
}

// NewServer creates a new server that can be started by calling Start.
func NewServer(address, port string) *Server {
	return &Server{
		address: address,
		port:    port,
	}
}

// SetOnServeErrorHandler sets the handler to call when an error occurs during
// serving.
func (s *Server) SetOnServeErrorHandler(onServeErrorHandler func(error)) {
	s.onServeErrorHandler = onServeErrorHandler
}

// Serve starts the HTTP server in a goroutine.
//
// Also functions as a force-restart by calling ForceStop if the server is
// already running, followed by attempting to launch it again.
//
// To stop the server you may use GracefulStop or ForceStop.
func (s *Server) Serve() {
	s.ForceStop()

	gin.DefaultWriter = ginutil.DefaultLoggerWriter
	gin.DefaultErrorWriter = ginutil.DefaultLoggerWriter

	r := gin.New()
	r.Use(
		ginutil.DefaultLoggerHandler,
		ginutil.RecoverProblem,
	)

	g := r.Group("/api")
	g.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	{
		buildModule := buildModule{}
		artifactModule := artifactModule{}
		buildModule.register(g)
		artifactModule.register(g)
	}

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.address, s.port),
		Handler: r,
	}

	go func() {
		log.Info().WithString("bindAddress", s.srv.Addr).Message("Server starting.")
		if err := s.srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info().Message("Closed server.")
			} else if s.onServeErrorHandler != nil {
				s.onServeErrorHandler(err)
			}
		}
	}()
}

// GracefulStop stops the HTTP server gracefully, blocking new connections
// and closing idle connections, then waiting until active ones have finished
// or the timeout duration has elapsed.
//
// If 0 is passed as the timeout, a default value of 30 seconds will be used.
func (s *Server) GracefulStop(timeout time.Duration) error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

// ForceStop forcefully stops the server.
func (s *Server) ForceStop() error {
	defer func() {
		s.isRunning = false
	}()
	if s.srv == nil {
		return nil
	}
	return s.srv.Close()
}

// IsRunning returns true if the server is currently running and processing
// requests.
func (s *Server) IsRunning() bool {
	return s.srv != nil && s.isRunning
}
